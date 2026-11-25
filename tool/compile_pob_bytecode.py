"""Собрать байткод мостов движка для той библиотеки LuaJIT, что запустит PoB."""

import argparse
import ctypes as c
import hashlib
import json
import shutil
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
BRIDGE = ROOT / "internal/pob/bridge"
LEGACY_LUA = ROOT / "internal/poblegacy/assets/lua"
EMBED = ROOT / "internal/pob/bytecode"
LUAJIT_MAGIC = b"\x1bLJ"
DUMP = b"local source, name = ...; return string.dump(assert(load(source, name, 't')), 'd')"


class Compiler:
    def __init__(self, library):
        self.lua = c.CDLL(str(library))
        self.lua.luaL_newstate.restype = c.c_void_p
        self.lua.luaL_openlibs.argtypes = (c.c_void_p,)
        self.lua.luaL_openlibs.restype = None
        self.lua.luaL_loadstring.argtypes = (c.c_void_p, c.c_char_p)
        self.lua.lua_pcall.argtypes = (c.c_void_p, c.c_int, c.c_int, c.c_int)
        self.lua.lua_pushlstring.argtypes = (c.c_void_p, c.c_char_p, c.c_size_t)
        self.lua.lua_pushlstring.restype = None
        self.lua.lua_tolstring.argtypes = (c.c_void_p, c.c_int, c.POINTER(c.c_size_t))
        self.lua.lua_tolstring.restype = c.c_void_p
        self.lua.lua_settop.argtypes = (c.c_void_p, c.c_int)
        self.lua.lua_settop.restype = None
        self.lua.lua_close.argtypes = (c.c_void_p,)
        self.lua.lua_close.restype = None
        self.state = self.lua.luaL_newstate()
        if not self.state:
            raise RuntimeError("LuaJIT не создал состояние компилятора")
        self.lua.luaL_openlibs(self.state)

    def close(self):
        self.lua.lua_close(self.state)

    def compile(self, source, name):
        if source.startswith(b"#"):
            source = b"\n" + source.partition(b"\n")[2]
        try:
            return self._dump(source, name)
        finally:
            self.lua.lua_settop(self.state, 0)

    def _dump(self, source, name):
        self._check(self.lua.luaL_loadstring(self.state, DUMP))
        name_bytes = ("@" + name).encode()
        self.lua.lua_pushlstring(self.state, source, len(source))
        self.lua.lua_pushlstring(self.state, name_bytes, len(name_bytes))
        self._check(self.lua.lua_pcall(self.state, 2, 1, 0))
        data = self._top_string()
        if not data.startswith(LUAJIT_MAGIC):
            raise ValueError(f"{name}: неверный заголовок байткода LuaJIT")
        return data

    def _check(self, status):
        if status:
            raise ValueError(self._top_string().decode("utf-8", errors="replace"))

    def _top_string(self):
        size = c.c_size_t()
        pointer = self.lua.lua_tolstring(self.state, -1, c.byref(size))
        return c.string_at(pointer, size.value) if pointer else b""


def compile_tree(compiler, source, target, prefix_entry=None):
    stats = {"files": 0, "source_bytes": 0, "bytecode_bytes": 0}
    for path in sorted(source.rglob("*.lua")):
        relative = path.relative_to(source)
        contents = path.read_bytes()
        if prefix_entry and relative.as_posix() == "headless_server.lua":
            contents = prefix_entry + b"\n" + contents
        data = compiler.compile(contents, relative.as_posix())
        output = target / relative
        output.parent.mkdir(parents=True, exist_ok=True)
        output.write_bytes(data)
        stats["files"] += 1
        stats["source_bytes"] += len(contents)
        stats["bytecode_bytes"] += len(data)
    return stats


def installation_library(manifest):
    current = json.loads(manifest.read_text(encoding="utf-8-sig"))["current"]
    directory = Path(current["directory"])
    runtime = directory / "runtime"
    if not (runtime / "lua/dkjson.lua").is_file():
        runtime = directory
    if sys.platform == "win32":
        return runtime / "lua51.dll"
    native = Path(current.get("nativeDirectory") or runtime)
    library = "libluajit-5.1.dylib" if sys.platform == "darwin" else "libluajit-5.1.so.2"
    return native / library


def prepare_bridges(compiler, library):
    output = EMBED / "embedded"
    output.mkdir(parents=True, exist_ok=True)
    # Остатки предыдущей генерации не должны попасть в новую сборку.
    for path in output.rglob("*.lua"):
        path.unlink()
    manifest = output / "manifest.json"
    manifest.unlink(missing_ok=True)

    result = {
        "bridge": compile_tree(compiler, BRIDGE, output / "bridge"),
        "workbench": compile_tree(
            compiler, LEGACY_LUA, output / "workbench/lua",
            prefix_entry=(BRIDGE / "workbench.lua").read_bytes(),
        ),
    }
    geometry = BRIDGE / "tree_geometry.lua"
    (output / "workbench/tree_geometry.lua").write_bytes(
        compiler.compile(geometry.read_bytes(), geometry.name),
    )
    metadata = {"schema": 1, "library_sha256": hashlib.sha256(library.read_bytes()).hexdigest()}
    manifest.write_text(json.dumps(metadata, indent=2) + "\n", encoding="utf-8", newline="\n")
    return result | metadata


def binary_tree_reader(copy):
    source = copy / "src" if (copy / "src/Launch.lua").is_file() else copy
    path = source / "Classes/PassiveTree.lua"
    opened_as_text = b'io.open("TreeData/"..treeVersion.."/tree.lua", "r")'
    opened_as_binary = b'io.open("TreeData/"..treeVersion.."/tree.lua", "rb")'
    contents = path.read_bytes()
    if contents.count(opened_as_binary) == 1:
        return
    if contents.count(opened_as_text) != 1:
        raise ValueError(f"{path}: загрузчик tree.lua отличается от PoB v0.23.1")
    path.write_bytes(contents.replace(opened_as_text, opened_as_binary))


def prepare(compiler, source, plain, output):
    if plain.is_relative_to(source) or output.is_relative_to(source):
        raise ValueError("Копии PoB нужно создать вне исходного каталога")
    if output.is_relative_to(plain) or plain.is_relative_to(output):
        raise ValueError("Копии PoB должны находиться в разных каталогах")
    shutil.copytree(source, plain)
    binary_tree_reader(plain)
    shutil.copytree(plain, output)
    return compile_tree(compiler, plain, output)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    runtime = parser.add_mutually_exclusive_group(required=True)
    runtime.add_argument("--library", type=Path, help="библиотека LuaJIT для текущей платформы")
    runtime.add_argument("--installation", type=Path, help="JSON из pob2-engine pob-data ensure")
    parser.add_argument("--pob", type=Path, help="эксперимент: исходная установка PoB")
    parser.add_argument("--source-copy", type=Path, help="эксперимент: новая контрольная копия PoB")
    parser.add_argument("--output", type=Path, help="эксперимент: новый каталог PoB с байткодом")
    args = parser.parse_args()
    experiment = (args.pob, args.source_copy, args.output)
    if any(experiment) and not all(experiment):
        parser.error("Для эксперимента задайте вместе --pob, --source-copy и --output")
    library = (args.library or installation_library(args.installation)).resolve(strict=True)
    compiler = Compiler(library)
    try:
        stats = prepare_bridges(compiler, library)
        if all(experiment):
            stats["pob"] = prepare(compiler, args.pob.resolve(strict=True), args.source_copy.resolve(), args.output.resolve())
    finally:
        compiler.close()
    print(json.dumps(stats, indent=2))


if __name__ == "__main__":
    main()
