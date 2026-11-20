// This is a generated file - do not edit.
//
// Generated from pob2/v1/build.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'build.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'build.pbenum.dart';

/// Лексическое значение сохраняется для точного повтора аргументов команды.
class ConfigValue extends $pb.GeneratedMessage {
  factory ConfigValue({
    ConfigValueType? type,
    $core.String? value,
  }) {
    final result = ConfigValue._();
    if (type != null) result.type = type;
    if (value != null) result.value = value;
    return result;
  }

  ConfigValue._();

  factory ConfigValue.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfigValue()..mergeFromBuffer(data, registry);
  factory ConfigValue.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfigValue()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConfigValue',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: ConfigValue.$_createMessage)
    ..aE<ConfigValueType>(1, _omitFieldNames ? '' : 'type',
        enumValues: ConfigValueType.values)
    ..aOS(2, _omitFieldNames ? '' : 'value')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfigValue clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfigValue copyWith(void Function(ConfigValue) updates) =>
      super.copyWith((message) => updates(message as ConfigValue))
          as ConfigValue;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ConfigValue() / ConfigValue.new instead')
  static ConfigValue create() => ConfigValue._();
  static $pb.GeneratedMessage $_createMessage() => ConfigValue._();
  @$core.override
  ConfigValue createEmptyInstance() => ConfigValue._();
  @$core.pragma('dart2js:noInline')
  static ConfigValue getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConfigValue>(
          ConfigValue.$_createMessage);
  static ConfigValue? _defaultInstance;

  @$pb.TagNumber(1)
  ConfigValueType get type => $_getN(0);
  @$pb.TagNumber(1)
  set type(ConfigValueType value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get value => $_getSZ(1);
  @$pb.TagNumber(2)
  set value($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => $_clearField(2);
}

class BuildIssue extends $pb.GeneratedMessage {
  factory BuildIssue({
    $core.String? code,
    $core.String? path,
    $core.String? message,
  }) {
    final result = BuildIssue._();
    if (code != null) result.code = code;
    if (path != null) result.path = path;
    if (message != null) result.message = message;
    return result;
  }

  BuildIssue._();

  factory BuildIssue.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildIssue()..mergeFromBuffer(data, registry);
  factory BuildIssue.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildIssue()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BuildIssue',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: BuildIssue.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'code')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..aOS(3, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildIssue clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildIssue copyWith(void Function(BuildIssue) updates) =>
      super.copyWith((message) => updates(message as BuildIssue)) as BuildIssue;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use BuildIssue() / BuildIssue.new instead')
  static BuildIssue create() => BuildIssue._();
  static $pb.GeneratedMessage $_createMessage() => BuildIssue._();
  @$core.override
  BuildIssue createEmptyInstance() => BuildIssue._();
  @$core.pragma('dart2js:noInline')
  static BuildIssue getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<BuildIssue>(BuildIssue.$_createMessage);
  static BuildIssue? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get code => $_getSZ(0);
  @$pb.TagNumber(1)
  set code($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get message => $_getSZ(2);
  @$pb.TagNumber(3)
  set message($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMessage() => $_has(2);
  @$pb.TagNumber(3)
  void clearMessage() => $_clearField(3);
}

class ActiveSets extends $pb.GeneratedMessage {
  factory ActiveSets({
    $core.int? skills,
    $core.int? items,
    $core.int? tree,
    $core.int? config,
  }) {
    final result = ActiveSets._();
    if (skills != null) result.skills = skills;
    if (items != null) result.items = items;
    if (tree != null) result.tree = tree;
    if (config != null) result.config = config;
    return result;
  }

  ActiveSets._();

  factory ActiveSets.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ActiveSets()..mergeFromBuffer(data, registry);
  factory ActiveSets.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ActiveSets()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ActiveSets',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: ActiveSets.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'skills', fieldType: $pb.PbFieldType.OU3)
    ..aI(2, _omitFieldNames ? '' : 'items', fieldType: $pb.PbFieldType.OU3)
    ..aI(3, _omitFieldNames ? '' : 'tree', fieldType: $pb.PbFieldType.OU3)
    ..aI(4, _omitFieldNames ? '' : 'config', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ActiveSets clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ActiveSets copyWith(void Function(ActiveSets) updates) =>
      super.copyWith((message) => updates(message as ActiveSets)) as ActiveSets;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ActiveSets() / ActiveSets.new instead')
  static ActiveSets create() => ActiveSets._();
  static $pb.GeneratedMessage $_createMessage() => ActiveSets._();
  @$core.override
  ActiveSets createEmptyInstance() => ActiveSets._();
  @$core.pragma('dart2js:noInline')
  static ActiveSets getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ActiveSets>(ActiveSets.$_createMessage);
  static ActiveSets? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get skills => $_getIZ(0);
  @$pb.TagNumber(1)
  set skills($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSkills() => $_has(0);
  @$pb.TagNumber(1)
  void clearSkills() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get items => $_getIZ(1);
  @$pb.TagNumber(2)
  set items($core.int value) => $_setUnsignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasItems() => $_has(1);
  @$pb.TagNumber(2)
  void clearItems() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get tree => $_getIZ(2);
  @$pb.TagNumber(3)
  set tree($core.int value) => $_setUnsignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTree() => $_has(2);
  @$pb.TagNumber(3)
  void clearTree() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get config => $_getIZ(3);
  @$pb.TagNumber(4)
  set config($core.int value) => $_setUnsignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasConfig() => $_has(3);
  @$pb.TagNumber(4)
  void clearConfig() => $_clearField(4);
}

class BuildGem extends $pb.GeneratedMessage {
  factory BuildGem({
    $core.String? id,
    $core.String? name,
    $core.String? skillId,
    $core.String? variantId,
    $core.int? level,
    $core.int? quality,
    $core.bool? enabled,
    $core.int? statSet,
    $core.bool? corrupted,
  }) {
    final result = BuildGem._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (skillId != null) result.skillId = skillId;
    if (variantId != null) result.variantId = variantId;
    if (level != null) result.level = level;
    if (quality != null) result.quality = quality;
    if (enabled != null) result.enabled = enabled;
    if (statSet != null) result.statSet = statSet;
    if (corrupted != null) result.corrupted = corrupted;
    return result;
  }

  BuildGem._();

  factory BuildGem.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildGem()..mergeFromBuffer(data, registry);
  factory BuildGem.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildGem()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BuildGem',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: BuildGem.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'skillId')
    ..aOS(4, _omitFieldNames ? '' : 'variantId')
    ..aI(5, _omitFieldNames ? '' : 'level', fieldType: $pb.PbFieldType.OU3)
    ..aI(6, _omitFieldNames ? '' : 'quality', fieldType: $pb.PbFieldType.OU3)
    ..aOB(7, _omitFieldNames ? '' : 'enabled')
    ..aI(8, _omitFieldNames ? '' : 'statSet', fieldType: $pb.PbFieldType.OU3)
    ..aOB(9, _omitFieldNames ? '' : 'corrupted')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildGem clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildGem copyWith(void Function(BuildGem) updates) =>
      super.copyWith((message) => updates(message as BuildGem)) as BuildGem;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use BuildGem() / BuildGem.new instead')
  static BuildGem create() => BuildGem._();
  static $pb.GeneratedMessage $_createMessage() => BuildGem._();
  @$core.override
  BuildGem createEmptyInstance() => BuildGem._();
  @$core.pragma('dart2js:noInline')
  static BuildGem getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<BuildGem>(BuildGem.$_createMessage);
  static BuildGem? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get skillId => $_getSZ(2);
  @$pb.TagNumber(3)
  set skillId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSkillId() => $_has(2);
  @$pb.TagNumber(3)
  void clearSkillId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get variantId => $_getSZ(3);
  @$pb.TagNumber(4)
  set variantId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasVariantId() => $_has(3);
  @$pb.TagNumber(4)
  void clearVariantId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get level => $_getIZ(4);
  @$pb.TagNumber(5)
  set level($core.int value) => $_setUnsignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasLevel() => $_has(4);
  @$pb.TagNumber(5)
  void clearLevel() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.int get quality => $_getIZ(5);
  @$pb.TagNumber(6)
  set quality($core.int value) => $_setUnsignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasQuality() => $_has(5);
  @$pb.TagNumber(6)
  void clearQuality() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.bool get enabled => $_getBF(6);
  @$pb.TagNumber(7)
  set enabled($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasEnabled() => $_has(6);
  @$pb.TagNumber(7)
  void clearEnabled() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.int get statSet => $_getIZ(7);
  @$pb.TagNumber(8)
  set statSet($core.int value) => $_setUnsignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasStatSet() => $_has(7);
  @$pb.TagNumber(8)
  void clearStatSet() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.bool get corrupted => $_getBF(8);
  @$pb.TagNumber(9)
  set corrupted($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(9)
  $core.bool hasCorrupted() => $_has(8);
  @$pb.TagNumber(9)
  void clearCorrupted() => $_clearField(9);
}

class SkillGroup extends $pb.GeneratedMessage {
  factory SkillGroup({
    $core.int? id,
    $core.String? label,
    $core.bool? enabled,
    $core.int? mainSkill,
    $core.Iterable<BuildGem>? gems,
  }) {
    final result = SkillGroup._();
    if (id != null) result.id = id;
    if (label != null) result.label = label;
    if (enabled != null) result.enabled = enabled;
    if (mainSkill != null) result.mainSkill = mainSkill;
    if (gems != null) result.gems.addAll(gems);
    return result;
  }

  SkillGroup._();

  factory SkillGroup.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SkillGroup()..mergeFromBuffer(data, registry);
  factory SkillGroup.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SkillGroup()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SkillGroup',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: SkillGroup.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'label')
    ..aOB(3, _omitFieldNames ? '' : 'enabled')
    ..aI(4, _omitFieldNames ? '' : 'mainSkill', fieldType: $pb.PbFieldType.OU3)
    ..pPM<BuildGem>(5, _omitFieldNames ? '' : 'gems',
        subBuilder: BuildGem.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SkillGroup clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SkillGroup copyWith(void Function(SkillGroup) updates) =>
      super.copyWith((message) => updates(message as SkillGroup)) as SkillGroup;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SkillGroup() / SkillGroup.new instead')
  static SkillGroup create() => SkillGroup._();
  static $pb.GeneratedMessage $_createMessage() => SkillGroup._();
  @$core.override
  SkillGroup createEmptyInstance() => SkillGroup._();
  @$core.pragma('dart2js:noInline')
  static SkillGroup getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SkillGroup>(SkillGroup.$_createMessage);
  static SkillGroup? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get label => $_getSZ(1);
  @$pb.TagNumber(2)
  set label($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLabel() => $_has(1);
  @$pb.TagNumber(2)
  void clearLabel() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get enabled => $_getBF(2);
  @$pb.TagNumber(3)
  set enabled($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEnabled() => $_has(2);
  @$pb.TagNumber(3)
  void clearEnabled() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get mainSkill => $_getIZ(3);
  @$pb.TagNumber(4)
  set mainSkill($core.int value) => $_setUnsignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMainSkill() => $_has(3);
  @$pb.TagNumber(4)
  void clearMainSkill() => $_clearField(4);

  @$pb.TagNumber(5)
  $pb.PbList<BuildGem> get gems => $_getList(4);
}

class BuildItem extends $pb.GeneratedMessage {
  factory BuildItem({
    $core.int? id,
    $core.String? raw,
  }) {
    final result = BuildItem._();
    if (id != null) result.id = id;
    if (raw != null) result.raw = raw;
    return result;
  }

  BuildItem._();

  factory BuildItem.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildItem()..mergeFromBuffer(data, registry);
  factory BuildItem.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildItem()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BuildItem',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: BuildItem.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'id', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'raw')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildItem clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildItem copyWith(void Function(BuildItem) updates) =>
      super.copyWith((message) => updates(message as BuildItem)) as BuildItem;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use BuildItem() / BuildItem.new instead')
  static BuildItem create() => BuildItem._();
  static $pb.GeneratedMessage $_createMessage() => BuildItem._();
  @$core.override
  BuildItem createEmptyInstance() => BuildItem._();
  @$core.pragma('dart2js:noInline')
  static BuildItem getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<BuildItem>(BuildItem.$_createMessage);
  static BuildItem? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get id => $_getIZ(0);
  @$pb.TagNumber(1)
  set id($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get raw => $_getSZ(1);
  @$pb.TagNumber(2)
  set raw($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRaw() => $_has(1);
  @$pb.TagNumber(2)
  void clearRaw() => $_clearField(2);
}

class BuildModel extends $pb.GeneratedMessage {
  factory BuildModel({
    $core.String? targetVersion,
    $core.String? treeVersion,
    $core.String? className,
    $core.int? classId,
    $core.String? ascendancy,
    $core.int? ascendancyId,
    $core.int? level,
    $core.int? mainGroup,
    $core.int? weaponSet,
    ActiveSets? activeSets,
    $core.Iterable<SkillGroup>? groups,
    $core.Iterable<BuildItem>? items,
    $core.Iterable<$core.MapEntry<$core.String, $core.int>>? slots,
    $core.Iterable<$core.int>? nodes,
    $core.Iterable<$core.MapEntry<$core.String, ConfigValue>>? config,
    $core.Iterable<BuildIssue>? issues,
  }) {
    final result = BuildModel._();
    if (targetVersion != null) result.targetVersion = targetVersion;
    if (treeVersion != null) result.treeVersion = treeVersion;
    if (className != null) result.className = className;
    if (classId != null) result.classId = classId;
    if (ascendancy != null) result.ascendancy = ascendancy;
    if (ascendancyId != null) result.ascendancyId = ascendancyId;
    if (level != null) result.level = level;
    if (mainGroup != null) result.mainGroup = mainGroup;
    if (weaponSet != null) result.weaponSet = weaponSet;
    if (activeSets != null) result.activeSets = activeSets;
    if (groups != null) result.groups.addAll(groups);
    if (items != null) result.items.addAll(items);
    if (slots != null) result.slots.addEntries(slots);
    if (nodes != null) result.nodes.addAll(nodes);
    if (config != null) result.config.addEntries(config);
    if (issues != null) result.issues.addAll(issues);
    return result;
  }

  BuildModel._();

  factory BuildModel.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildModel()..mergeFromBuffer(data, registry);
  factory BuildModel.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildModel()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BuildModel',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: BuildModel.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'targetVersion')
    ..aOS(2, _omitFieldNames ? '' : 'treeVersion')
    ..aOS(3, _omitFieldNames ? '' : 'className')
    ..aI(4, _omitFieldNames ? '' : 'classId', fieldType: $pb.PbFieldType.OU3)
    ..aOS(5, _omitFieldNames ? '' : 'ascendancy')
    ..aI(6, _omitFieldNames ? '' : 'ascendancyId',
        fieldType: $pb.PbFieldType.OU3)
    ..aI(7, _omitFieldNames ? '' : 'level', fieldType: $pb.PbFieldType.OU3)
    ..aI(8, _omitFieldNames ? '' : 'mainGroup', fieldType: $pb.PbFieldType.OU3)
    ..aI(9, _omitFieldNames ? '' : 'weaponSet', fieldType: $pb.PbFieldType.OU3)
    ..aOM<ActiveSets>(10, _omitFieldNames ? '' : 'activeSets',
        subBuilder: ActiveSets.$_createMessage)
    ..pPM<SkillGroup>(11, _omitFieldNames ? '' : 'groups',
        subBuilder: SkillGroup.$_createMessage)
    ..pPM<BuildItem>(12, _omitFieldNames ? '' : 'items',
        subBuilder: BuildItem.$_createMessage)
    ..m<$core.String, $core.int>(13, _omitFieldNames ? '' : 'slots',
        entryClassName: 'BuildModel.SlotsEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OU3,
        packageName: const $pb.PackageName('pob2.v1'))
    ..p<$core.int>(14, _omitFieldNames ? '' : 'nodes', $pb.PbFieldType.KU3)
    ..m<$core.String, ConfigValue>(15, _omitFieldNames ? '' : 'config',
        entryClassName: 'BuildModel.ConfigEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OM,
        valueCreator: ConfigValue.$_createMessage,
        valueDefaultOrMaker: ConfigValue.getDefault,
        packageName: const $pb.PackageName('pob2.v1'))
    ..pPM<BuildIssue>(16, _omitFieldNames ? '' : 'issues',
        subBuilder: BuildIssue.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildModel clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildModel copyWith(void Function(BuildModel) updates) =>
      super.copyWith((message) => updates(message as BuildModel)) as BuildModel;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use BuildModel() / BuildModel.new instead')
  static BuildModel create() => BuildModel._();
  static $pb.GeneratedMessage $_createMessage() => BuildModel._();
  @$core.override
  BuildModel createEmptyInstance() => BuildModel._();
  @$core.pragma('dart2js:noInline')
  static BuildModel getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<BuildModel>(BuildModel.$_createMessage);
  static BuildModel? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get targetVersion => $_getSZ(0);
  @$pb.TagNumber(1)
  set targetVersion($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTargetVersion() => $_has(0);
  @$pb.TagNumber(1)
  void clearTargetVersion() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get treeVersion => $_getSZ(1);
  @$pb.TagNumber(2)
  set treeVersion($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTreeVersion() => $_has(1);
  @$pb.TagNumber(2)
  void clearTreeVersion() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get className => $_getSZ(2);
  @$pb.TagNumber(3)
  set className($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasClassName() => $_has(2);
  @$pb.TagNumber(3)
  void clearClassName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get classId => $_getIZ(3);
  @$pb.TagNumber(4)
  set classId($core.int value) => $_setUnsignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasClassId() => $_has(3);
  @$pb.TagNumber(4)
  void clearClassId() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get ascendancy => $_getSZ(4);
  @$pb.TagNumber(5)
  set ascendancy($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAscendancy() => $_has(4);
  @$pb.TagNumber(5)
  void clearAscendancy() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.int get ascendancyId => $_getIZ(5);
  @$pb.TagNumber(6)
  set ascendancyId($core.int value) => $_setUnsignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAscendancyId() => $_has(5);
  @$pb.TagNumber(6)
  void clearAscendancyId() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.int get level => $_getIZ(6);
  @$pb.TagNumber(7)
  set level($core.int value) => $_setUnsignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasLevel() => $_has(6);
  @$pb.TagNumber(7)
  void clearLevel() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.int get mainGroup => $_getIZ(7);
  @$pb.TagNumber(8)
  set mainGroup($core.int value) => $_setUnsignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasMainGroup() => $_has(7);
  @$pb.TagNumber(8)
  void clearMainGroup() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.int get weaponSet => $_getIZ(8);
  @$pb.TagNumber(9)
  set weaponSet($core.int value) => $_setUnsignedInt32(8, value);
  @$pb.TagNumber(9)
  $core.bool hasWeaponSet() => $_has(8);
  @$pb.TagNumber(9)
  void clearWeaponSet() => $_clearField(9);

  @$pb.TagNumber(10)
  ActiveSets get activeSets => $_getN(9);
  @$pb.TagNumber(10)
  set activeSets(ActiveSets value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasActiveSets() => $_has(9);
  @$pb.TagNumber(10)
  void clearActiveSets() => $_clearField(10);
  @$pb.TagNumber(10)
  ActiveSets ensureActiveSets() => $_ensure(9);

  @$pb.TagNumber(11)
  $pb.PbList<SkillGroup> get groups => $_getList(10);

  @$pb.TagNumber(12)
  $pb.PbList<BuildItem> get items => $_getList(11);

  @$pb.TagNumber(13)
  $pb.PbMap<$core.String, $core.int> get slots => $_getMap(12);

  @$pb.TagNumber(14)
  $pb.PbList<$core.int> get nodes => $_getList(13);

  @$pb.TagNumber(15)
  $pb.PbMap<$core.String, ConfigValue> get config => $_getMap(14);

  @$pb.TagNumber(16)
  $pb.PbList<BuildIssue> get issues => $_getList(15);
}

class AppliedModifier extends $pb.GeneratedMessage {
  factory AppliedModifier({
    $core.String? stat,
    $core.String? kind,
    $core.double? value,
    $core.String? source,
  }) {
    final result = AppliedModifier._();
    if (stat != null) result.stat = stat;
    if (kind != null) result.kind = kind;
    if (value != null) result.value = value;
    if (source != null) result.source = source;
    return result;
  }

  AppliedModifier._();

  factory AppliedModifier.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AppliedModifier()..mergeFromBuffer(data, registry);
  factory AppliedModifier.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AppliedModifier()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AppliedModifier',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: AppliedModifier.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'stat')
    ..aOS(2, _omitFieldNames ? '' : 'kind')
    ..aD(3, _omitFieldNames ? '' : 'value')
    ..aOS(4, _omitFieldNames ? '' : 'source')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AppliedModifier clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AppliedModifier copyWith(void Function(AppliedModifier) updates) =>
      super.copyWith((message) => updates(message as AppliedModifier))
          as AppliedModifier;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AppliedModifier() / AppliedModifier.new instead')
  static AppliedModifier create() => AppliedModifier._();
  static $pb.GeneratedMessage $_createMessage() => AppliedModifier._();
  @$core.override
  AppliedModifier createEmptyInstance() => AppliedModifier._();
  @$core.pragma('dart2js:noInline')
  static AppliedModifier getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AppliedModifier>(
          AppliedModifier.$_createMessage);
  static AppliedModifier? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get stat => $_getSZ(0);
  @$pb.TagNumber(1)
  set stat($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasStat() => $_has(0);
  @$pb.TagNumber(1)
  void clearStat() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get kind => $_getSZ(1);
  @$pb.TagNumber(2)
  set kind($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasKind() => $_has(1);
  @$pb.TagNumber(2)
  void clearKind() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.double get value => $_getN(2);
  @$pb.TagNumber(3)
  set value($core.double value) => $_setDouble(2, value);
  @$pb.TagNumber(3)
  $core.bool hasValue() => $_has(2);
  @$pb.TagNumber(3)
  void clearValue() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get source => $_getSZ(3);
  @$pb.TagNumber(4)
  set source($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSource() => $_has(3);
  @$pb.TagNumber(4)
  void clearSource() => $_clearField(4);
}

class CalculationResult extends $pb.GeneratedMessage {
  factory CalculationResult({
    CalculationStatus? status,
    $core.String? dataId,
    $core.String? dataSha256,
    $core.String? pobCommit,
    $core.String? inputSha256,
    $core.String? skill,
    $core.String? mode,
    $core.Iterable<$core.MapEntry<$core.String, $core.double>>? stats,
    $core.Iterable<AppliedModifier>? modifiers,
    $core.Iterable<BuildIssue>? issues,
  }) {
    final result = CalculationResult._();
    if (status != null) result.status = status;
    if (dataId != null) result.dataId = dataId;
    if (dataSha256 != null) result.dataSha256 = dataSha256;
    if (pobCommit != null) result.pobCommit = pobCommit;
    if (inputSha256 != null) result.inputSha256 = inputSha256;
    if (skill != null) result.skill = skill;
    if (mode != null) result.mode = mode;
    if (stats != null) result.stats.addEntries(stats);
    if (modifiers != null) result.modifiers.addAll(modifiers);
    if (issues != null) result.issues.addAll(issues);
    return result;
  }

  CalculationResult._();

  factory CalculationResult.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CalculationResult()..mergeFromBuffer(data, registry);
  factory CalculationResult.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CalculationResult()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CalculationResult',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: CalculationResult.$_createMessage)
    ..aE<CalculationStatus>(1, _omitFieldNames ? '' : 'status',
        enumValues: CalculationStatus.values)
    ..aOS(2, _omitFieldNames ? '' : 'dataId')
    ..aOS(3, _omitFieldNames ? '' : 'dataSha256')
    ..aOS(4, _omitFieldNames ? '' : 'pobCommit')
    ..aOS(5, _omitFieldNames ? '' : 'inputSha256')
    ..aOS(6, _omitFieldNames ? '' : 'skill')
    ..aOS(7, _omitFieldNames ? '' : 'mode')
    ..m<$core.String, $core.double>(8, _omitFieldNames ? '' : 'stats',
        entryClassName: 'CalculationResult.StatsEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OD,
        packageName: const $pb.PackageName('pob2.v1'))
    ..pPM<AppliedModifier>(9, _omitFieldNames ? '' : 'modifiers',
        subBuilder: AppliedModifier.$_createMessage)
    ..pPM<BuildIssue>(10, _omitFieldNames ? '' : 'issues',
        subBuilder: BuildIssue.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CalculationResult clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CalculationResult copyWith(void Function(CalculationResult) updates) =>
      super.copyWith((message) => updates(message as CalculationResult))
          as CalculationResult;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CalculationResult() / CalculationResult.new instead')
  static CalculationResult create() => CalculationResult._();
  static $pb.GeneratedMessage $_createMessage() => CalculationResult._();
  @$core.override
  CalculationResult createEmptyInstance() => CalculationResult._();
  @$core.pragma('dart2js:noInline')
  static CalculationResult getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CalculationResult>(
          CalculationResult.$_createMessage);
  static CalculationResult? _defaultInstance;

  @$pb.TagNumber(1)
  CalculationStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(CalculationStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get dataId => $_getSZ(1);
  @$pb.TagNumber(2)
  set dataId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDataId() => $_has(1);
  @$pb.TagNumber(2)
  void clearDataId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get dataSha256 => $_getSZ(2);
  @$pb.TagNumber(3)
  set dataSha256($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDataSha256() => $_has(2);
  @$pb.TagNumber(3)
  void clearDataSha256() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get pobCommit => $_getSZ(3);
  @$pb.TagNumber(4)
  set pobCommit($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPobCommit() => $_has(3);
  @$pb.TagNumber(4)
  void clearPobCommit() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get inputSha256 => $_getSZ(4);
  @$pb.TagNumber(5)
  set inputSha256($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasInputSha256() => $_has(4);
  @$pb.TagNumber(5)
  void clearInputSha256() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get skill => $_getSZ(5);
  @$pb.TagNumber(6)
  set skill($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasSkill() => $_has(5);
  @$pb.TagNumber(6)
  void clearSkill() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get mode => $_getSZ(6);
  @$pb.TagNumber(7)
  set mode($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasMode() => $_has(6);
  @$pb.TagNumber(7)
  void clearMode() => $_clearField(7);

  @$pb.TagNumber(8)
  $pb.PbMap<$core.String, $core.double> get stats => $_getMap(7);

  @$pb.TagNumber(9)
  $pb.PbList<AppliedModifier> get modifiers => $_getList(8);

  @$pb.TagNumber(10)
  $pb.PbList<BuildIssue> get issues => $_getList(9);
}

class BuildDocument extends $pb.GeneratedMessage {
  factory BuildDocument({
    $core.String? xml,
    $core.String? inputSha256,
    $fixnum.Int64? inputRevision,
    BuildModel? model,
    CalculationResult? calculation,
    $fixnum.Int64? calculatedFromRevision,
  }) {
    final result = BuildDocument._();
    if (xml != null) result.xml = xml;
    if (inputSha256 != null) result.inputSha256 = inputSha256;
    if (inputRevision != null) result.inputRevision = inputRevision;
    if (model != null) result.model = model;
    if (calculation != null) result.calculation = calculation;
    if (calculatedFromRevision != null)
      result.calculatedFromRevision = calculatedFromRevision;
    return result;
  }

  BuildDocument._();

  factory BuildDocument.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildDocument()..mergeFromBuffer(data, registry);
  factory BuildDocument.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildDocument()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BuildDocument',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: BuildDocument.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'xml')
    ..aOS(2, _omitFieldNames ? '' : 'inputSha256')
    ..a<$fixnum.Int64>(
        3, _omitFieldNames ? '' : 'inputRevision', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOM<BuildModel>(4, _omitFieldNames ? '' : 'model',
        subBuilder: BuildModel.$_createMessage)
    ..aOM<CalculationResult>(5, _omitFieldNames ? '' : 'calculation',
        subBuilder: CalculationResult.$_createMessage)
    ..a<$fixnum.Int64>(
        6, _omitFieldNames ? '' : 'calculatedFromRevision', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildDocument clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildDocument copyWith(void Function(BuildDocument) updates) =>
      super.copyWith((message) => updates(message as BuildDocument))
          as BuildDocument;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use BuildDocument() / BuildDocument.new instead')
  static BuildDocument create() => BuildDocument._();
  static $pb.GeneratedMessage $_createMessage() => BuildDocument._();
  @$core.override
  BuildDocument createEmptyInstance() => BuildDocument._();
  @$core.pragma('dart2js:noInline')
  static BuildDocument getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<BuildDocument>(
          BuildDocument.$_createMessage);
  static BuildDocument? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get xml => $_getSZ(0);
  @$pb.TagNumber(1)
  set xml($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasXml() => $_has(0);
  @$pb.TagNumber(1)
  void clearXml() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get inputSha256 => $_getSZ(1);
  @$pb.TagNumber(2)
  set inputSha256($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInputSha256() => $_has(1);
  @$pb.TagNumber(2)
  void clearInputSha256() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get inputRevision => $_getI64(2);
  @$pb.TagNumber(3)
  set inputRevision($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasInputRevision() => $_has(2);
  @$pb.TagNumber(3)
  void clearInputRevision() => $_clearField(3);

  @$pb.TagNumber(4)
  BuildModel get model => $_getN(3);
  @$pb.TagNumber(4)
  set model(BuildModel value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasModel() => $_has(3);
  @$pb.TagNumber(4)
  void clearModel() => $_clearField(4);
  @$pb.TagNumber(4)
  BuildModel ensureModel() => $_ensure(3);

  @$pb.TagNumber(5)
  CalculationResult get calculation => $_getN(4);
  @$pb.TagNumber(5)
  set calculation(CalculationResult value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasCalculation() => $_has(4);
  @$pb.TagNumber(5)
  void clearCalculation() => $_clearField(5);
  @$pb.TagNumber(5)
  CalculationResult ensureCalculation() => $_ensure(4);

  @$pb.TagNumber(6)
  $fixnum.Int64 get calculatedFromRevision => $_getI64(5);
  @$pb.TagNumber(6)
  set calculatedFromRevision($fixnum.Int64 value) => $_setInt64(5, value);
  @$pb.TagNumber(6)
  $core.bool hasCalculatedFromRevision() => $_has(5);
  @$pb.TagNumber(6)
  void clearCalculatedFromRevision() => $_clearField(6);
}

class LevelEdit extends $pb.GeneratedMessage {
  factory LevelEdit({
    $core.int? level,
  }) {
    final result = LevelEdit._();
    if (level != null) result.level = level;
    return result;
  }

  LevelEdit._();

  factory LevelEdit.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LevelEdit()..mergeFromBuffer(data, registry);
  factory LevelEdit.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LevelEdit()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LevelEdit',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: LevelEdit.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'level', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LevelEdit clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LevelEdit copyWith(void Function(LevelEdit) updates) =>
      super.copyWith((message) => updates(message as LevelEdit)) as LevelEdit;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LevelEdit() / LevelEdit.new instead')
  static LevelEdit create() => LevelEdit._();
  static $pb.GeneratedMessage $_createMessage() => LevelEdit._();
  @$core.override
  LevelEdit createEmptyInstance() => LevelEdit._();
  @$core.pragma('dart2js:noInline')
  static LevelEdit getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LevelEdit>(LevelEdit.$_createMessage);
  static LevelEdit? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get level => $_getIZ(0);
  @$pb.TagNumber(1)
  set level($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasLevel() => $_has(0);
  @$pb.TagNumber(1)
  void clearLevel() => $_clearField(1);
}

class ConfigEdit extends $pb.GeneratedMessage {
  factory ConfigEdit({
    $core.String? name,
    ConfigValue? value,
  }) {
    final result = ConfigEdit._();
    if (name != null) result.name = name;
    if (value != null) result.value = value;
    return result;
  }

  ConfigEdit._();

  factory ConfigEdit.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfigEdit()..mergeFromBuffer(data, registry);
  factory ConfigEdit.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfigEdit()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConfigEdit',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: ConfigEdit.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOM<ConfigValue>(2, _omitFieldNames ? '' : 'value',
        subBuilder: ConfigValue.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfigEdit clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfigEdit copyWith(void Function(ConfigEdit) updates) =>
      super.copyWith((message) => updates(message as ConfigEdit)) as ConfigEdit;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ConfigEdit() / ConfigEdit.new instead')
  static ConfigEdit create() => ConfigEdit._();
  static $pb.GeneratedMessage $_createMessage() => ConfigEdit._();
  @$core.override
  ConfigEdit createEmptyInstance() => ConfigEdit._();
  @$core.pragma('dart2js:noInline')
  static ConfigEdit getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ConfigEdit>(ConfigEdit.$_createMessage);
  static ConfigEdit? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Отсутствие value удаляет условие.
  @$pb.TagNumber(2)
  ConfigValue get value => $_getN(1);
  @$pb.TagNumber(2)
  set value(ConfigValue value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => $_clearField(2);
  @$pb.TagNumber(2)
  ConfigValue ensureValue() => $_ensure(1);
}

class GemInput extends $pb.GeneratedMessage {
  factory GemInput({
    $core.String? id,
    $core.int? level,
    $core.int? quality,
    $core.bool? enabled,
  }) {
    final result = GemInput._();
    if (id != null) result.id = id;
    if (level != null) result.level = level;
    if (quality != null) result.quality = quality;
    if (enabled != null) result.enabled = enabled;
    return result;
  }

  GemInput._();

  factory GemInput.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GemInput()..mergeFromBuffer(data, registry);
  factory GemInput.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GemInput()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GemInput',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: GemInput.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aI(2, _omitFieldNames ? '' : 'level', fieldType: $pb.PbFieldType.OU3)
    ..aI(3, _omitFieldNames ? '' : 'quality', fieldType: $pb.PbFieldType.OU3)
    ..aOB(4, _omitFieldNames ? '' : 'enabled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GemInput clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GemInput copyWith(void Function(GemInput) updates) =>
      super.copyWith((message) => updates(message as GemInput)) as GemInput;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GemInput() / GemInput.new instead')
  static GemInput create() => GemInput._();
  static $pb.GeneratedMessage $_createMessage() => GemInput._();
  @$core.override
  GemInput createEmptyInstance() => GemInput._();
  @$core.pragma('dart2js:noInline')
  static GemInput getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GemInput>(GemInput.$_createMessage);
  static GemInput? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get level => $_getIZ(1);
  @$pb.TagNumber(2)
  set level($core.int value) => $_setUnsignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLevel() => $_has(1);
  @$pb.TagNumber(2)
  void clearLevel() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get quality => $_getIZ(2);
  @$pb.TagNumber(3)
  set quality($core.int value) => $_setUnsignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasQuality() => $_has(2);
  @$pb.TagNumber(3)
  void clearQuality() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get enabled => $_getBF(3);
  @$pb.TagNumber(4)
  set enabled($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEnabled() => $_has(3);
  @$pb.TagNumber(4)
  void clearEnabled() => $_clearField(4);
}

class GemsEdit extends $pb.GeneratedMessage {
  factory GemsEdit({
    $core.int? groupId,
    $core.String? label,
    $core.bool? enabled,
    $core.bool? delete,
    $core.Iterable<GemInput>? gems,
  }) {
    final result = GemsEdit._();
    if (groupId != null) result.groupId = groupId;
    if (label != null) result.label = label;
    if (enabled != null) result.enabled = enabled;
    if (delete != null) result.delete = delete;
    if (gems != null) result.gems.addAll(gems);
    return result;
  }

  GemsEdit._();

  factory GemsEdit.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GemsEdit()..mergeFromBuffer(data, registry);
  factory GemsEdit.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GemsEdit()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GemsEdit',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: GemsEdit.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'groupId', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'label')
    ..aOB(3, _omitFieldNames ? '' : 'enabled')
    ..aOB(4, _omitFieldNames ? '' : 'delete')
    ..pPM<GemInput>(5, _omitFieldNames ? '' : 'gems',
        subBuilder: GemInput.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GemsEdit clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GemsEdit copyWith(void Function(GemsEdit) updates) =>
      super.copyWith((message) => updates(message as GemsEdit)) as GemsEdit;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GemsEdit() / GemsEdit.new instead')
  static GemsEdit create() => GemsEdit._();
  static $pb.GeneratedMessage $_createMessage() => GemsEdit._();
  @$core.override
  GemsEdit createEmptyInstance() => GemsEdit._();
  @$core.pragma('dart2js:noInline')
  static GemsEdit getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GemsEdit>(GemsEdit.$_createMessage);
  static GemsEdit? _defaultInstance;

  /// Ноль добавляет группу в конец выбранного набора.
  @$pb.TagNumber(1)
  $core.int get groupId => $_getIZ(0);
  @$pb.TagNumber(1)
  set groupId($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasGroupId() => $_has(0);
  @$pb.TagNumber(1)
  void clearGroupId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get label => $_getSZ(1);
  @$pb.TagNumber(2)
  set label($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLabel() => $_has(1);
  @$pb.TagNumber(2)
  void clearLabel() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get enabled => $_getBF(2);
  @$pb.TagNumber(3)
  set enabled($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEnabled() => $_has(2);
  @$pb.TagNumber(3)
  void clearEnabled() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get delete => $_getBF(3);
  @$pb.TagNumber(4)
  set delete($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDelete() => $_has(3);
  @$pb.TagNumber(4)
  void clearDelete() => $_clearField(4);

  @$pb.TagNumber(5)
  $pb.PbList<GemInput> get gems => $_getList(4);
}

class EquipmentEdit extends $pb.GeneratedMessage {
  factory EquipmentEdit({
    $core.String? slot,
    $core.String? raw,
  }) {
    final result = EquipmentEdit._();
    if (slot != null) result.slot = slot;
    if (raw != null) result.raw = raw;
    return result;
  }

  EquipmentEdit._();

  factory EquipmentEdit.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EquipmentEdit()..mergeFromBuffer(data, registry);
  factory EquipmentEdit.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      EquipmentEdit()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EquipmentEdit',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: EquipmentEdit.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'slot')
    ..aOS(2, _omitFieldNames ? '' : 'raw')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EquipmentEdit clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EquipmentEdit copyWith(void Function(EquipmentEdit) updates) =>
      super.copyWith((message) => updates(message as EquipmentEdit))
          as EquipmentEdit;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use EquipmentEdit() / EquipmentEdit.new instead')
  static EquipmentEdit create() => EquipmentEdit._();
  static $pb.GeneratedMessage $_createMessage() => EquipmentEdit._();
  @$core.override
  EquipmentEdit createEmptyInstance() => EquipmentEdit._();
  @$core.pragma('dart2js:noInline')
  static EquipmentEdit getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EquipmentEdit>(
          EquipmentEdit.$_createMessage);
  static EquipmentEdit? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get slot => $_getSZ(0);
  @$pb.TagNumber(1)
  set slot($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSlot() => $_has(0);
  @$pb.TagNumber(1)
  void clearSlot() => $_clearField(1);

  /// Пустая строка снимает предмет.
  @$pb.TagNumber(2)
  $core.String get raw => $_getSZ(1);
  @$pb.TagNumber(2)
  set raw($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRaw() => $_has(1);
  @$pb.TagNumber(2)
  void clearRaw() => $_clearField(2);
}

class PassivesEdit extends $pb.GeneratedMessage {
  factory PassivesEdit({
    $core.Iterable<$core.int>? nodes,
  }) {
    final result = PassivesEdit._();
    if (nodes != null) result.nodes.addAll(nodes);
    return result;
  }

  PassivesEdit._();

  factory PassivesEdit.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PassivesEdit()..mergeFromBuffer(data, registry);
  factory PassivesEdit.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PassivesEdit()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PassivesEdit',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: PassivesEdit.$_createMessage)
    ..p<$core.int>(1, _omitFieldNames ? '' : 'nodes', $pb.PbFieldType.KU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PassivesEdit clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PassivesEdit copyWith(void Function(PassivesEdit) updates) =>
      super.copyWith((message) => updates(message as PassivesEdit))
          as PassivesEdit;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PassivesEdit() / PassivesEdit.new instead')
  static PassivesEdit create() => PassivesEdit._();
  static $pb.GeneratedMessage $_createMessage() => PassivesEdit._();
  @$core.override
  PassivesEdit createEmptyInstance() => PassivesEdit._();
  @$core.pragma('dart2js:noInline')
  static PassivesEdit getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PassivesEdit>(
          PassivesEdit.$_createMessage);
  static PassivesEdit? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.int> get nodes => $_getList(0);
}

class SelectionEdit extends $pb.GeneratedMessage {
  factory SelectionEdit({
    ActiveSets? sets,
    $core.int? mainGroup,
    $core.int? weaponSet,
  }) {
    final result = SelectionEdit._();
    if (sets != null) result.sets = sets;
    if (mainGroup != null) result.mainGroup = mainGroup;
    if (weaponSet != null) result.weaponSet = weaponSet;
    return result;
  }

  SelectionEdit._();

  factory SelectionEdit.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SelectionEdit()..mergeFromBuffer(data, registry);
  factory SelectionEdit.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SelectionEdit()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SelectionEdit',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: SelectionEdit.$_createMessage)
    ..aOM<ActiveSets>(1, _omitFieldNames ? '' : 'sets',
        subBuilder: ActiveSets.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'mainGroup', fieldType: $pb.PbFieldType.OU3)
    ..aI(3, _omitFieldNames ? '' : 'weaponSet', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SelectionEdit clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SelectionEdit copyWith(void Function(SelectionEdit) updates) =>
      super.copyWith((message) => updates(message as SelectionEdit))
          as SelectionEdit;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SelectionEdit() / SelectionEdit.new instead')
  static SelectionEdit create() => SelectionEdit._();
  static $pb.GeneratedMessage $_createMessage() => SelectionEdit._();
  @$core.override
  SelectionEdit createEmptyInstance() => SelectionEdit._();
  @$core.pragma('dart2js:noInline')
  static SelectionEdit getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SelectionEdit>(
          SelectionEdit.$_createMessage);
  static SelectionEdit? _defaultInstance;

  @$pb.TagNumber(1)
  ActiveSets get sets => $_getN(0);
  @$pb.TagNumber(1)
  set sets(ActiveSets value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSets() => $_has(0);
  @$pb.TagNumber(1)
  void clearSets() => $_clearField(1);
  @$pb.TagNumber(1)
  ActiveSets ensureSets() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.int get mainGroup => $_getIZ(1);
  @$pb.TagNumber(2)
  set mainGroup($core.int value) => $_setUnsignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMainGroup() => $_has(1);
  @$pb.TagNumber(2)
  void clearMainGroup() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get weaponSet => $_getIZ(2);
  @$pb.TagNumber(3)
  set weaponSet($core.int value) => $_setUnsignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasWeaponSet() => $_has(2);
  @$pb.TagNumber(3)
  void clearWeaponSet() => $_clearField(3);
}

enum BuildEdit_Operation {
  level,
  config,
  gems,
  equipment,
  passives,
  selection,
  notSet
}

class BuildEdit extends $pb.GeneratedMessage {
  factory BuildEdit({
    LevelEdit? level,
    ConfigEdit? config,
    GemsEdit? gems,
    EquipmentEdit? equipment,
    PassivesEdit? passives,
    SelectionEdit? selection,
  }) {
    final result = BuildEdit._();
    if (level != null) result.level = level;
    if (config != null) result.config = config;
    if (gems != null) result.gems = gems;
    if (equipment != null) result.equipment = equipment;
    if (passives != null) result.passives = passives;
    if (selection != null) result.selection = selection;
    return result;
  }

  BuildEdit._();

  factory BuildEdit.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildEdit()..mergeFromBuffer(data, registry);
  factory BuildEdit.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildEdit()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, BuildEdit_Operation>
      _BuildEdit_OperationByTag = {
    1: BuildEdit_Operation.level,
    2: BuildEdit_Operation.config,
    3: BuildEdit_Operation.gems,
    4: BuildEdit_Operation.equipment,
    5: BuildEdit_Operation.passives,
    6: BuildEdit_Operation.selection,
    0: BuildEdit_Operation.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BuildEdit',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: BuildEdit.$_createMessage)
    ..oo(0, [1, 2, 3, 4, 5, 6])
    ..aOM<LevelEdit>(1, _omitFieldNames ? '' : 'level',
        subBuilder: LevelEdit.$_createMessage)
    ..aOM<ConfigEdit>(2, _omitFieldNames ? '' : 'config',
        subBuilder: ConfigEdit.$_createMessage)
    ..aOM<GemsEdit>(3, _omitFieldNames ? '' : 'gems',
        subBuilder: GemsEdit.$_createMessage)
    ..aOM<EquipmentEdit>(4, _omitFieldNames ? '' : 'equipment',
        subBuilder: EquipmentEdit.$_createMessage)
    ..aOM<PassivesEdit>(5, _omitFieldNames ? '' : 'passives',
        subBuilder: PassivesEdit.$_createMessage)
    ..aOM<SelectionEdit>(6, _omitFieldNames ? '' : 'selection',
        subBuilder: SelectionEdit.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildEdit clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildEdit copyWith(void Function(BuildEdit) updates) =>
      super.copyWith((message) => updates(message as BuildEdit)) as BuildEdit;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use BuildEdit() / BuildEdit.new instead')
  static BuildEdit create() => BuildEdit._();
  static $pb.GeneratedMessage $_createMessage() => BuildEdit._();
  @$core.override
  BuildEdit createEmptyInstance() => BuildEdit._();
  @$core.pragma('dart2js:noInline')
  static BuildEdit getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<BuildEdit>(BuildEdit.$_createMessage);
  static BuildEdit? _defaultInstance;

  @$pb.TagNumber(1)
  @$pb.TagNumber(2)
  @$pb.TagNumber(3)
  @$pb.TagNumber(4)
  @$pb.TagNumber(5)
  @$pb.TagNumber(6)
  BuildEdit_Operation whichOperation() =>
      _BuildEdit_OperationByTag[$_whichOneof(0)]!;
  @$pb.TagNumber(1)
  @$pb.TagNumber(2)
  @$pb.TagNumber(3)
  @$pb.TagNumber(4)
  @$pb.TagNumber(5)
  @$pb.TagNumber(6)
  void clearOperation() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  LevelEdit get level => $_getN(0);
  @$pb.TagNumber(1)
  set level(LevelEdit value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasLevel() => $_has(0);
  @$pb.TagNumber(1)
  void clearLevel() => $_clearField(1);
  @$pb.TagNumber(1)
  LevelEdit ensureLevel() => $_ensure(0);

  @$pb.TagNumber(2)
  ConfigEdit get config => $_getN(1);
  @$pb.TagNumber(2)
  set config(ConfigEdit value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasConfig() => $_has(1);
  @$pb.TagNumber(2)
  void clearConfig() => $_clearField(2);
  @$pb.TagNumber(2)
  ConfigEdit ensureConfig() => $_ensure(1);

  @$pb.TagNumber(3)
  GemsEdit get gems => $_getN(2);
  @$pb.TagNumber(3)
  set gems(GemsEdit value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasGems() => $_has(2);
  @$pb.TagNumber(3)
  void clearGems() => $_clearField(3);
  @$pb.TagNumber(3)
  GemsEdit ensureGems() => $_ensure(2);

  @$pb.TagNumber(4)
  EquipmentEdit get equipment => $_getN(3);
  @$pb.TagNumber(4)
  set equipment(EquipmentEdit value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasEquipment() => $_has(3);
  @$pb.TagNumber(4)
  void clearEquipment() => $_clearField(4);
  @$pb.TagNumber(4)
  EquipmentEdit ensureEquipment() => $_ensure(3);

  @$pb.TagNumber(5)
  PassivesEdit get passives => $_getN(4);
  @$pb.TagNumber(5)
  set passives(PassivesEdit value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasPassives() => $_has(4);
  @$pb.TagNumber(5)
  void clearPassives() => $_clearField(5);
  @$pb.TagNumber(5)
  PassivesEdit ensurePassives() => $_ensure(4);

  @$pb.TagNumber(6)
  SelectionEdit get selection => $_getN(5);
  @$pb.TagNumber(6)
  set selection(SelectionEdit value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasSelection() => $_has(5);
  @$pb.TagNumber(6)
  void clearSelection() => $_clearField(6);
  @$pb.TagNumber(6)
  SelectionEdit ensureSelection() => $_ensure(5);
}

class StatDelta extends $pb.GeneratedMessage {
  factory StatDelta({
    $core.String? stat,
    $core.double? before,
    $core.double? after,
    $core.double? difference,
  }) {
    final result = StatDelta._();
    if (stat != null) result.stat = stat;
    if (before != null) result.before = before;
    if (after != null) result.after = after;
    if (difference != null) result.difference = difference;
    return result;
  }

  StatDelta._();

  factory StatDelta.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StatDelta()..mergeFromBuffer(data, registry);
  factory StatDelta.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      StatDelta()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StatDelta',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: StatDelta.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'stat')
    ..aD(2, _omitFieldNames ? '' : 'before')
    ..aD(3, _omitFieldNames ? '' : 'after')
    ..aD(4, _omitFieldNames ? '' : 'difference')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StatDelta clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StatDelta copyWith(void Function(StatDelta) updates) =>
      super.copyWith((message) => updates(message as StatDelta)) as StatDelta;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use StatDelta() / StatDelta.new instead')
  static StatDelta create() => StatDelta._();
  static $pb.GeneratedMessage $_createMessage() => StatDelta._();
  @$core.override
  StatDelta createEmptyInstance() => StatDelta._();
  @$core.pragma('dart2js:noInline')
  static StatDelta getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StatDelta>(StatDelta.$_createMessage);
  static StatDelta? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get stat => $_getSZ(0);
  @$pb.TagNumber(1)
  set stat($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasStat() => $_has(0);
  @$pb.TagNumber(1)
  void clearStat() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.double get before => $_getN(1);
  @$pb.TagNumber(2)
  set before($core.double value) => $_setDouble(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBefore() => $_has(1);
  @$pb.TagNumber(2)
  void clearBefore() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.double get after => $_getN(2);
  @$pb.TagNumber(3)
  set after($core.double value) => $_setDouble(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAfter() => $_has(2);
  @$pb.TagNumber(3)
  void clearAfter() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.double get difference => $_getN(3);
  @$pb.TagNumber(4)
  set difference($core.double value) => $_setDouble(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDifference() => $_has(3);
  @$pb.TagNumber(4)
  void clearDifference() => $_clearField(4);
}

class BuildComparison extends $pb.GeneratedMessage {
  factory BuildComparison({
    CalculationResult? before,
    CalculationResult? after,
    $core.Iterable<StatDelta>? deltas,
  }) {
    final result = BuildComparison._();
    if (before != null) result.before = before;
    if (after != null) result.after = after;
    if (deltas != null) result.deltas.addAll(deltas);
    return result;
  }

  BuildComparison._();

  factory BuildComparison.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildComparison()..mergeFromBuffer(data, registry);
  factory BuildComparison.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BuildComparison()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BuildComparison',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: BuildComparison.$_createMessage)
    ..aOM<CalculationResult>(1, _omitFieldNames ? '' : 'before',
        subBuilder: CalculationResult.$_createMessage)
    ..aOM<CalculationResult>(2, _omitFieldNames ? '' : 'after',
        subBuilder: CalculationResult.$_createMessage)
    ..pPM<StatDelta>(3, _omitFieldNames ? '' : 'deltas',
        subBuilder: StatDelta.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildComparison clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BuildComparison copyWith(void Function(BuildComparison) updates) =>
      super.copyWith((message) => updates(message as BuildComparison))
          as BuildComparison;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use BuildComparison() / BuildComparison.new instead')
  static BuildComparison create() => BuildComparison._();
  static $pb.GeneratedMessage $_createMessage() => BuildComparison._();
  @$core.override
  BuildComparison createEmptyInstance() => BuildComparison._();
  @$core.pragma('dart2js:noInline')
  static BuildComparison getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<BuildComparison>(
          BuildComparison.$_createMessage);
  static BuildComparison? _defaultInstance;

  @$pb.TagNumber(1)
  CalculationResult get before => $_getN(0);
  @$pb.TagNumber(1)
  set before(CalculationResult value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBefore() => $_has(0);
  @$pb.TagNumber(1)
  void clearBefore() => $_clearField(1);
  @$pb.TagNumber(1)
  CalculationResult ensureBefore() => $_ensure(0);

  @$pb.TagNumber(2)
  CalculationResult get after => $_getN(1);
  @$pb.TagNumber(2)
  set after(CalculationResult value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasAfter() => $_has(1);
  @$pb.TagNumber(2)
  void clearAfter() => $_clearField(2);
  @$pb.TagNumber(2)
  CalculationResult ensureAfter() => $_ensure(1);

  /// Пуст при неподдержанном варианте; это не нулевые значения показателей.
  @$pb.TagNumber(3)
  $pb.PbList<StatDelta> get deltas => $_getList(2);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
