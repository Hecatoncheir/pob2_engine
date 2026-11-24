-- Данные одного кадра для изображений; формулы и состояния остаются в PoB.
return function(methods)
    local json = require("dkjson")
    local readTreeGeometry = dofile(POB_MCP_PACKAGE_DIR .. "/tree_geometry.lua")
    local baseCommands, baseRevision
    local function commandKey(command)
        return json.encode({command.file, command.index, command.positions, command.uv, command.color, command.blend, command.clamp})
    end
    local function captureTree(hoverId)
        local capture = dofile(POB_MCP_PACKAGE_DIR .. "/lua/tree_scene.lua")
        setfenv(capture, getfenv(1))
        return capture(build, {treeScope = "full", width = 1856, height = 2034, hoverId = hoverId})
    end
    local function imageTree()
        local geometry = readTreeGeometry(build.spec)
        local nodes, edges = {}, {}
        local hidden = setmetatable({}, { __jsontype = "array" })
        for _, node in ipairs(geometry.nodes) do
            nodes[#nodes + 1] = { id = node.id, x = node.x, y = node.y, name = node.name,
                type = node.kind, allocated = node.allocated, isAttribute = node.attribute,
                icon = node.icon, drawSize = node.draw_size, allocMode = node.allocation_mode }
        end
        for _, edge in ipairs(geometry.edges) do
            local pair = {edge.from_node, edge.to_node}
            edges[#edges + 1] = pair
            if edge.hidden_in_scene then hidden[#hidden + 1] = pair end
        end
        return { nodes = nodes, edges = edges, hiddenEdges = hidden }
    end

    function methods.getBuildImageData(params)
        local result = { report = methods.getBuildReport() }
        if params.view == "tree" then
            result.tree = imageTree()
            local capture = dofile(POB_MCP_PACKAGE_DIR .. "/lua/tree_scene.lua")
            setfenv(capture, getfenv(1))
            result.scene = capture(build, params)
        end
        if params.view == "node" then
            result.tooltip = methods.getNodeTooltip({ id = params.id, compare = false })
        end
        return result
    end

    function methods.getTreeGeometry(params)
        local result = imageTree()
        result.version = build.spec.treeVersion
        result.pointDisplay = build.controls.pointDisplay.str
        if params and params.includeScene then
            result.scene = captureTree()
            baseCommands, baseRevision = {}, params.revision
            for _, command in ipairs(result.scene.commands) do baseCommands[commandKey(command)] = true end
        end
        return result
    end

    function methods.getTreeHover(params)
        build.spec:BuildAllDependsAndPaths()
        if not baseCommands or baseRevision ~= params.revision then
            methods.getTreeGeometry({includeScene = true, revision = params.revision})
        end
        local scene = captureTree(params.id)
        local tooltip = methods.getNodeTooltip({id = params.id, compare = params.compare ~= false})
        local overlay = {}
        for _, command in ipairs(scene.commands) do
            if not baseCommands[commandKey(command)] then overlay[#overlay + 1] = command end
        end
        scene.commands = overlay
        return {id = params.id, tooltip = tooltip, scene = scene}
    end

    function methods.getTreeIconAssets(params)
        local result = {}
        local known = {}
        for _, node in pairs(build.spec.nodes) do
            if node.icon then known[node.icon] = true end
        end
        for _, key in ipairs(params.icons or {}) do
            local asset = known[key] and build.spec.tree:GetAssetByName(key)
            if asset and asset.handle and asset.handle.sourcePath then
                local uv = asset[2] and {asset[1], asset[2], asset[3], asset[4]} or {0, 0, 1, 1}
                result[#result + 1] = {key = key, file = asset.handle.sourcePath,
                    index = asset[2] and (asset[5] or 1) or (asset[1] or 1), uv = uv}
            else
                result[#result + 1] = {key = key, error = "Иконка отсутствует в исходном дереве PoB"}
            end
        end
        return {icons = result}
    end
end
