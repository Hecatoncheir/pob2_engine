// This is a generated file - do not edit.
//
// Generated from pob2/v1/build.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use configValueTypeDescriptor instead')
const ConfigValueType$json = {
  '1': 'ConfigValueType',
  '2': [
    {'1': 'CONFIG_VALUE_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'CONFIG_VALUE_TYPE_BOOLEAN', '2': 1},
    {'1': 'CONFIG_VALUE_TYPE_NUMBER', '2': 2},
    {'1': 'CONFIG_VALUE_TYPE_STRING', '2': 3},
  ],
};

/// Descriptor for `ConfigValueType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List configValueTypeDescriptor = $convert.base64Decode(
    'Cg9Db25maWdWYWx1ZVR5cGUSIQodQ09ORklHX1ZBTFVFX1RZUEVfVU5TUEVDSUZJRUQQABIdCh'
    'lDT05GSUdfVkFMVUVfVFlQRV9CT09MRUFOEAESHAoYQ09ORklHX1ZBTFVFX1RZUEVfTlVNQkVS'
    'EAISHAoYQ09ORklHX1ZBTFVFX1RZUEVfU1RSSU5HEAM=');

@$core.Deprecated('Use calculationStatusDescriptor instead')
const CalculationStatus$json = {
  '1': 'CalculationStatus',
  '2': [
    {'1': 'CALCULATION_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'CALCULATION_STATUS_COMPLETE', '2': 1},
    {'1': 'CALCULATION_STATUS_PARTIAL', '2': 2},
    {'1': 'CALCULATION_STATUS_UNSUPPORTED', '2': 3},
  ],
};

/// Descriptor for `CalculationStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List calculationStatusDescriptor = $convert.base64Decode(
    'ChFDYWxjdWxhdGlvblN0YXR1cxIiCh5DQUxDVUxBVElPTl9TVEFUVVNfVU5TUEVDSUZJRUQQAB'
    'IfChtDQUxDVUxBVElPTl9TVEFUVVNfQ09NUExFVEUQARIeChpDQUxDVUxBVElPTl9TVEFUVVNf'
    'UEFSVElBTBACEiIKHkNBTENVTEFUSU9OX1NUQVRVU19VTlNVUFBPUlRFRBAD');

@$core.Deprecated('Use configValueDescriptor instead')
const ConfigValue$json = {
  '1': 'ConfigValue',
  '2': [
    {
      '1': 'type',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pob2.v1.ConfigValueType',
      '10': 'type'
    },
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
};

/// Descriptor for `ConfigValue`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configValueDescriptor = $convert.base64Decode(
    'CgtDb25maWdWYWx1ZRIsCgR0eXBlGAEgASgOMhgucG9iMi52MS5Db25maWdWYWx1ZVR5cGVSBH'
    'R5cGUSFAoFdmFsdWUYAiABKAlSBXZhbHVl');

@$core.Deprecated('Use buildIssueDescriptor instead')
const BuildIssue$json = {
  '1': 'BuildIssue',
  '2': [
    {'1': 'code', '3': 1, '4': 1, '5': 9, '10': 'code'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '10': 'path'},
    {'1': 'message', '3': 3, '4': 1, '5': 9, '10': 'message'},
  ],
};

/// Descriptor for `BuildIssue`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List buildIssueDescriptor = $convert.base64Decode(
    'CgpCdWlsZElzc3VlEhIKBGNvZGUYASABKAlSBGNvZGUSEgoEcGF0aBgCIAEoCVIEcGF0aBIYCg'
    'dtZXNzYWdlGAMgASgJUgdtZXNzYWdl');

@$core.Deprecated('Use activeSetsDescriptor instead')
const ActiveSets$json = {
  '1': 'ActiveSets',
  '2': [
    {'1': 'skills', '3': 1, '4': 1, '5': 13, '10': 'skills'},
    {'1': 'items', '3': 2, '4': 1, '5': 13, '10': 'items'},
    {'1': 'tree', '3': 3, '4': 1, '5': 13, '10': 'tree'},
    {'1': 'config', '3': 4, '4': 1, '5': 13, '10': 'config'},
  ],
};

/// Descriptor for `ActiveSets`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List activeSetsDescriptor = $convert.base64Decode(
    'CgpBY3RpdmVTZXRzEhYKBnNraWxscxgBIAEoDVIGc2tpbGxzEhQKBWl0ZW1zGAIgASgNUgVpdG'
    'VtcxISCgR0cmVlGAMgASgNUgR0cmVlEhYKBmNvbmZpZxgEIAEoDVIGY29uZmln');

@$core.Deprecated('Use buildGemDescriptor instead')
const BuildGem$json = {
  '1': 'BuildGem',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'skill_id', '3': 3, '4': 1, '5': 9, '10': 'skillId'},
    {'1': 'variant_id', '3': 4, '4': 1, '5': 9, '10': 'variantId'},
    {'1': 'level', '3': 5, '4': 1, '5': 13, '10': 'level'},
    {'1': 'quality', '3': 6, '4': 1, '5': 13, '10': 'quality'},
    {'1': 'enabled', '3': 7, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'stat_set', '3': 8, '4': 1, '5': 13, '10': 'statSet'},
    {'1': 'corrupted', '3': 9, '4': 1, '5': 8, '10': 'corrupted'},
  ],
};

/// Descriptor for `BuildGem`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List buildGemDescriptor = $convert.base64Decode(
    'CghCdWlsZEdlbRIOCgJpZBgBIAEoCVICaWQSEgoEbmFtZRgCIAEoCVIEbmFtZRIZCghza2lsbF'
    '9pZBgDIAEoCVIHc2tpbGxJZBIdCgp2YXJpYW50X2lkGAQgASgJUgl2YXJpYW50SWQSFAoFbGV2'
    'ZWwYBSABKA1SBWxldmVsEhgKB3F1YWxpdHkYBiABKA1SB3F1YWxpdHkSGAoHZW5hYmxlZBgHIA'
    'EoCFIHZW5hYmxlZBIZCghzdGF0X3NldBgIIAEoDVIHc3RhdFNldBIcCgljb3JydXB0ZWQYCSAB'
    'KAhSCWNvcnJ1cHRlZA==');

@$core.Deprecated('Use skillGroupDescriptor instead')
const SkillGroup$json = {
  '1': 'SkillGroup',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'label', '3': 2, '4': 1, '5': 9, '10': 'label'},
    {'1': 'enabled', '3': 3, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'main_skill', '3': 4, '4': 1, '5': 13, '10': 'mainSkill'},
    {
      '1': 'gems',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.BuildGem',
      '10': 'gems'
    },
  ],
};

/// Descriptor for `SkillGroup`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List skillGroupDescriptor = $convert.base64Decode(
    'CgpTa2lsbEdyb3VwEg4KAmlkGAEgASgNUgJpZBIUCgVsYWJlbBgCIAEoCVIFbGFiZWwSGAoHZW'
    '5hYmxlZBgDIAEoCFIHZW5hYmxlZBIdCgptYWluX3NraWxsGAQgASgNUgltYWluU2tpbGwSJQoE'
    'Z2VtcxgFIAMoCzIRLnBvYjIudjEuQnVpbGRHZW1SBGdlbXM=');

@$core.Deprecated('Use buildItemDescriptor instead')
const BuildItem$json = {
  '1': 'BuildItem',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 13, '10': 'id'},
    {'1': 'raw', '3': 2, '4': 1, '5': 9, '10': 'raw'},
  ],
};

/// Descriptor for `BuildItem`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List buildItemDescriptor = $convert.base64Decode(
    'CglCdWlsZEl0ZW0SDgoCaWQYASABKA1SAmlkEhAKA3JhdxgCIAEoCVIDcmF3');

@$core.Deprecated('Use buildModelDescriptor instead')
const BuildModel$json = {
  '1': 'BuildModel',
  '2': [
    {'1': 'target_version', '3': 1, '4': 1, '5': 9, '10': 'targetVersion'},
    {'1': 'tree_version', '3': 2, '4': 1, '5': 9, '10': 'treeVersion'},
    {'1': 'class_name', '3': 3, '4': 1, '5': 9, '10': 'className'},
    {'1': 'class_id', '3': 4, '4': 1, '5': 13, '10': 'classId'},
    {'1': 'ascendancy', '3': 5, '4': 1, '5': 9, '10': 'ascendancy'},
    {'1': 'ascendancy_id', '3': 6, '4': 1, '5': 13, '10': 'ascendancyId'},
    {'1': 'level', '3': 7, '4': 1, '5': 13, '10': 'level'},
    {'1': 'main_group', '3': 8, '4': 1, '5': 13, '10': 'mainGroup'},
    {'1': 'weapon_set', '3': 9, '4': 1, '5': 13, '10': 'weaponSet'},
    {
      '1': 'active_sets',
      '3': 10,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.ActiveSets',
      '10': 'activeSets'
    },
    {
      '1': 'groups',
      '3': 11,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.SkillGroup',
      '10': 'groups'
    },
    {
      '1': 'items',
      '3': 12,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.BuildItem',
      '10': 'items'
    },
    {
      '1': 'slots',
      '3': 13,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.BuildModel.SlotsEntry',
      '10': 'slots'
    },
    {'1': 'nodes', '3': 14, '4': 3, '5': 13, '10': 'nodes'},
    {
      '1': 'config',
      '3': 15,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.BuildModel.ConfigEntry',
      '10': 'config'
    },
    {
      '1': 'issues',
      '3': 16,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.BuildIssue',
      '10': 'issues'
    },
  ],
  '3': [BuildModel_SlotsEntry$json, BuildModel_ConfigEntry$json],
};

@$core.Deprecated('Use buildModelDescriptor instead')
const BuildModel_SlotsEntry$json = {
  '1': 'SlotsEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 13, '10': 'value'},
  ],
  '7': {'7': true},
};

@$core.Deprecated('Use buildModelDescriptor instead')
const BuildModel_ConfigEntry$json = {
  '1': 'ConfigEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {
      '1': 'value',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.ConfigValue',
      '10': 'value'
    },
  ],
  '7': {'7': true},
};

/// Descriptor for `BuildModel`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List buildModelDescriptor = $convert.base64Decode(
    'CgpCdWlsZE1vZGVsEiUKDnRhcmdldF92ZXJzaW9uGAEgASgJUg10YXJnZXRWZXJzaW9uEiEKDH'
    'RyZWVfdmVyc2lvbhgCIAEoCVILdHJlZVZlcnNpb24SHQoKY2xhc3NfbmFtZRgDIAEoCVIJY2xh'
    'c3NOYW1lEhkKCGNsYXNzX2lkGAQgASgNUgdjbGFzc0lkEh4KCmFzY2VuZGFuY3kYBSABKAlSCm'
    'FzY2VuZGFuY3kSIwoNYXNjZW5kYW5jeV9pZBgGIAEoDVIMYXNjZW5kYW5jeUlkEhQKBWxldmVs'
    'GAcgASgNUgVsZXZlbBIdCgptYWluX2dyb3VwGAggASgNUgltYWluR3JvdXASHQoKd2VhcG9uX3'
    'NldBgJIAEoDVIJd2VhcG9uU2V0EjQKC2FjdGl2ZV9zZXRzGAogASgLMhMucG9iMi52MS5BY3Rp'
    'dmVTZXRzUgphY3RpdmVTZXRzEisKBmdyb3VwcxgLIAMoCzITLnBvYjIudjEuU2tpbGxHcm91cF'
    'IGZ3JvdXBzEigKBWl0ZW1zGAwgAygLMhIucG9iMi52MS5CdWlsZEl0ZW1SBWl0ZW1zEjQKBXNs'
    'b3RzGA0gAygLMh4ucG9iMi52MS5CdWlsZE1vZGVsLlNsb3RzRW50cnlSBXNsb3RzEhQKBW5vZG'
    'VzGA4gAygNUgVub2RlcxI3CgZjb25maWcYDyADKAsyHy5wb2IyLnYxLkJ1aWxkTW9kZWwuQ29u'
    'ZmlnRW50cnlSBmNvbmZpZxIrCgZpc3N1ZXMYECADKAsyEy5wb2IyLnYxLkJ1aWxkSXNzdWVSBm'
    'lzc3Vlcxo4CgpTbG90c0VudHJ5EhAKA2tleRgBIAEoCVIDa2V5EhQKBXZhbHVlGAIgASgNUgV2'
    'YWx1ZToCOAEaTwoLQ29uZmlnRW50cnkSEAoDa2V5GAEgASgJUgNrZXkSKgoFdmFsdWUYAiABKA'
    'syFC5wb2IyLnYxLkNvbmZpZ1ZhbHVlUgV2YWx1ZToCOAE=');

@$core.Deprecated('Use appliedModifierDescriptor instead')
const AppliedModifier$json = {
  '1': 'AppliedModifier',
  '2': [
    {'1': 'stat', '3': 1, '4': 1, '5': 9, '10': 'stat'},
    {'1': 'kind', '3': 2, '4': 1, '5': 9, '10': 'kind'},
    {'1': 'value', '3': 3, '4': 1, '5': 1, '10': 'value'},
    {'1': 'source', '3': 4, '4': 1, '5': 9, '10': 'source'},
  ],
};

/// Descriptor for `AppliedModifier`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List appliedModifierDescriptor = $convert.base64Decode(
    'Cg9BcHBsaWVkTW9kaWZpZXISEgoEc3RhdBgBIAEoCVIEc3RhdBISCgRraW5kGAIgASgJUgRraW'
    '5kEhQKBXZhbHVlGAMgASgBUgV2YWx1ZRIWCgZzb3VyY2UYBCABKAlSBnNvdXJjZQ==');

@$core.Deprecated('Use calculationResultDescriptor instead')
const CalculationResult$json = {
  '1': 'CalculationResult',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.pob2.v1.CalculationStatus',
      '10': 'status'
    },
    {'1': 'data_id', '3': 2, '4': 1, '5': 9, '10': 'dataId'},
    {'1': 'data_sha256', '3': 3, '4': 1, '5': 9, '10': 'dataSha256'},
    {'1': 'pob_commit', '3': 4, '4': 1, '5': 9, '10': 'pobCommit'},
    {'1': 'input_sha256', '3': 5, '4': 1, '5': 9, '10': 'inputSha256'},
    {'1': 'skill', '3': 6, '4': 1, '5': 9, '10': 'skill'},
    {'1': 'mode', '3': 7, '4': 1, '5': 9, '10': 'mode'},
    {
      '1': 'stats',
      '3': 8,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.CalculationResult.StatsEntry',
      '10': 'stats'
    },
    {
      '1': 'modifiers',
      '3': 9,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.AppliedModifier',
      '10': 'modifiers'
    },
    {
      '1': 'issues',
      '3': 10,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.BuildIssue',
      '10': 'issues'
    },
  ],
  '3': [CalculationResult_StatsEntry$json],
};

@$core.Deprecated('Use calculationResultDescriptor instead')
const CalculationResult_StatsEntry$json = {
  '1': 'StatsEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 1, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `CalculationResult`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List calculationResultDescriptor = $convert.base64Decode(
    'ChFDYWxjdWxhdGlvblJlc3VsdBIyCgZzdGF0dXMYASABKA4yGi5wb2IyLnYxLkNhbGN1bGF0aW'
    '9uU3RhdHVzUgZzdGF0dXMSFwoHZGF0YV9pZBgCIAEoCVIGZGF0YUlkEh8KC2RhdGFfc2hhMjU2'
    'GAMgASgJUgpkYXRhU2hhMjU2Eh0KCnBvYl9jb21taXQYBCABKAlSCXBvYkNvbW1pdBIhCgxpbn'
    'B1dF9zaGEyNTYYBSABKAlSC2lucHV0U2hhMjU2EhQKBXNraWxsGAYgASgJUgVza2lsbBISCgRt'
    'b2RlGAcgASgJUgRtb2RlEjsKBXN0YXRzGAggAygLMiUucG9iMi52MS5DYWxjdWxhdGlvblJlc3'
    'VsdC5TdGF0c0VudHJ5UgVzdGF0cxI2Cgltb2RpZmllcnMYCSADKAsyGC5wb2IyLnYxLkFwcGxp'
    'ZWRNb2RpZmllclIJbW9kaWZpZXJzEisKBmlzc3VlcxgKIAMoCzITLnBvYjIudjEuQnVpbGRJc3'
    'N1ZVIGaXNzdWVzGjgKClN0YXRzRW50cnkSEAoDa2V5GAEgASgJUgNrZXkSFAoFdmFsdWUYAiAB'
    'KAFSBXZhbHVlOgI4AQ==');

@$core.Deprecated('Use buildDocumentDescriptor instead')
const BuildDocument$json = {
  '1': 'BuildDocument',
  '2': [
    {'1': 'xml', '3': 1, '4': 1, '5': 9, '10': 'xml'},
    {'1': 'input_sha256', '3': 2, '4': 1, '5': 9, '10': 'inputSha256'},
    {'1': 'input_revision', '3': 3, '4': 1, '5': 4, '10': 'inputRevision'},
    {
      '1': 'model',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.BuildModel',
      '10': 'model'
    },
    {
      '1': 'calculation',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.CalculationResult',
      '10': 'calculation'
    },
    {
      '1': 'calculated_from_revision',
      '3': 6,
      '4': 1,
      '5': 4,
      '10': 'calculatedFromRevision'
    },
  ],
};

/// Descriptor for `BuildDocument`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List buildDocumentDescriptor = $convert.base64Decode(
    'Cg1CdWlsZERvY3VtZW50EhAKA3htbBgBIAEoCVIDeG1sEiEKDGlucHV0X3NoYTI1NhgCIAEoCV'
    'ILaW5wdXRTaGEyNTYSJQoOaW5wdXRfcmV2aXNpb24YAyABKARSDWlucHV0UmV2aXNpb24SKQoF'
    'bW9kZWwYBCABKAsyEy5wb2IyLnYxLkJ1aWxkTW9kZWxSBW1vZGVsEjwKC2NhbGN1bGF0aW9uGA'
    'UgASgLMhoucG9iMi52MS5DYWxjdWxhdGlvblJlc3VsdFILY2FsY3VsYXRpb24SOAoYY2FsY3Vs'
    'YXRlZF9mcm9tX3JldmlzaW9uGAYgASgEUhZjYWxjdWxhdGVkRnJvbVJldmlzaW9u');

@$core.Deprecated('Use levelEditDescriptor instead')
const LevelEdit$json = {
  '1': 'LevelEdit',
  '2': [
    {'1': 'level', '3': 1, '4': 1, '5': 13, '10': 'level'},
  ],
};

/// Descriptor for `LevelEdit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List levelEditDescriptor =
    $convert.base64Decode('CglMZXZlbEVkaXQSFAoFbGV2ZWwYASABKA1SBWxldmVs');

@$core.Deprecated('Use configEditDescriptor instead')
const ConfigEdit$json = {
  '1': 'ConfigEdit',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {
      '1': 'value',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.ConfigValue',
      '10': 'value'
    },
  ],
};

/// Descriptor for `ConfigEdit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configEditDescriptor = $convert.base64Decode(
    'CgpDb25maWdFZGl0EhIKBG5hbWUYASABKAlSBG5hbWUSKgoFdmFsdWUYAiABKAsyFC5wb2IyLn'
    'YxLkNvbmZpZ1ZhbHVlUgV2YWx1ZQ==');

@$core.Deprecated('Use gemInputDescriptor instead')
const GemInput$json = {
  '1': 'GemInput',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'level', '3': 2, '4': 1, '5': 13, '10': 'level'},
    {'1': 'quality', '3': 3, '4': 1, '5': 13, '10': 'quality'},
    {'1': 'enabled', '3': 4, '4': 1, '5': 8, '10': 'enabled'},
  ],
};

/// Descriptor for `GemInput`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List gemInputDescriptor = $convert.base64Decode(
    'CghHZW1JbnB1dBIOCgJpZBgBIAEoCVICaWQSFAoFbGV2ZWwYAiABKA1SBWxldmVsEhgKB3F1YW'
    'xpdHkYAyABKA1SB3F1YWxpdHkSGAoHZW5hYmxlZBgEIAEoCFIHZW5hYmxlZA==');

@$core.Deprecated('Use gemsEditDescriptor instead')
const GemsEdit$json = {
  '1': 'GemsEdit',
  '2': [
    {'1': 'group_id', '3': 1, '4': 1, '5': 13, '10': 'groupId'},
    {'1': 'label', '3': 2, '4': 1, '5': 9, '10': 'label'},
    {'1': 'enabled', '3': 3, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'delete', '3': 4, '4': 1, '5': 8, '10': 'delete'},
    {
      '1': 'gems',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.GemInput',
      '10': 'gems'
    },
  ],
};

/// Descriptor for `GemsEdit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List gemsEditDescriptor = $convert.base64Decode(
    'CghHZW1zRWRpdBIZCghncm91cF9pZBgBIAEoDVIHZ3JvdXBJZBIUCgVsYWJlbBgCIAEoCVIFbG'
    'FiZWwSGAoHZW5hYmxlZBgDIAEoCFIHZW5hYmxlZBIWCgZkZWxldGUYBCABKAhSBmRlbGV0ZRIl'
    'CgRnZW1zGAUgAygLMhEucG9iMi52MS5HZW1JbnB1dFIEZ2Vtcw==');

@$core.Deprecated('Use equipmentEditDescriptor instead')
const EquipmentEdit$json = {
  '1': 'EquipmentEdit',
  '2': [
    {'1': 'slot', '3': 1, '4': 1, '5': 9, '10': 'slot'},
    {'1': 'raw', '3': 2, '4': 1, '5': 9, '10': 'raw'},
  ],
};

/// Descriptor for `EquipmentEdit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List equipmentEditDescriptor = $convert.base64Decode(
    'Cg1FcXVpcG1lbnRFZGl0EhIKBHNsb3QYASABKAlSBHNsb3QSEAoDcmF3GAIgASgJUgNyYXc=');

@$core.Deprecated('Use passivesEditDescriptor instead')
const PassivesEdit$json = {
  '1': 'PassivesEdit',
  '2': [
    {'1': 'nodes', '3': 1, '4': 3, '5': 13, '10': 'nodes'},
  ],
};

/// Descriptor for `PassivesEdit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List passivesEditDescriptor =
    $convert.base64Decode('CgxQYXNzaXZlc0VkaXQSFAoFbm9kZXMYASADKA1SBW5vZGVz');

@$core.Deprecated('Use selectionEditDescriptor instead')
const SelectionEdit$json = {
  '1': 'SelectionEdit',
  '2': [
    {
      '1': 'sets',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.ActiveSets',
      '10': 'sets'
    },
    {'1': 'main_group', '3': 2, '4': 1, '5': 13, '10': 'mainGroup'},
    {'1': 'weapon_set', '3': 3, '4': 1, '5': 13, '10': 'weaponSet'},
  ],
};

/// Descriptor for `SelectionEdit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List selectionEditDescriptor = $convert.base64Decode(
    'Cg1TZWxlY3Rpb25FZGl0EicKBHNldHMYASABKAsyEy5wb2IyLnYxLkFjdGl2ZVNldHNSBHNldH'
    'MSHQoKbWFpbl9ncm91cBgCIAEoDVIJbWFpbkdyb3VwEh0KCndlYXBvbl9zZXQYAyABKA1SCXdl'
    'YXBvblNldA==');

@$core.Deprecated('Use buildEditDescriptor instead')
const BuildEdit$json = {
  '1': 'BuildEdit',
  '2': [
    {
      '1': 'level',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.LevelEdit',
      '9': 0,
      '10': 'level'
    },
    {
      '1': 'config',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.ConfigEdit',
      '9': 0,
      '10': 'config'
    },
    {
      '1': 'gems',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.GemsEdit',
      '9': 0,
      '10': 'gems'
    },
    {
      '1': 'equipment',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.EquipmentEdit',
      '9': 0,
      '10': 'equipment'
    },
    {
      '1': 'passives',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.PassivesEdit',
      '9': 0,
      '10': 'passives'
    },
    {
      '1': 'selection',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.SelectionEdit',
      '9': 0,
      '10': 'selection'
    },
  ],
  '8': [
    {'1': 'operation'},
  ],
};

/// Descriptor for `BuildEdit`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List buildEditDescriptor = $convert.base64Decode(
    'CglCdWlsZEVkaXQSKgoFbGV2ZWwYASABKAsyEi5wb2IyLnYxLkxldmVsRWRpdEgAUgVsZXZlbB'
    'ItCgZjb25maWcYAiABKAsyEy5wb2IyLnYxLkNvbmZpZ0VkaXRIAFIGY29uZmlnEicKBGdlbXMY'
    'AyABKAsyES5wb2IyLnYxLkdlbXNFZGl0SABSBGdlbXMSNgoJZXF1aXBtZW50GAQgASgLMhYucG'
    '9iMi52MS5FcXVpcG1lbnRFZGl0SABSCWVxdWlwbWVudBIzCghwYXNzaXZlcxgFIAEoCzIVLnBv'
    'YjIudjEuUGFzc2l2ZXNFZGl0SABSCHBhc3NpdmVzEjYKCXNlbGVjdGlvbhgGIAEoCzIWLnBvYj'
    'IudjEuU2VsZWN0aW9uRWRpdEgAUglzZWxlY3Rpb25CCwoJb3BlcmF0aW9u');

@$core.Deprecated('Use statDeltaDescriptor instead')
const StatDelta$json = {
  '1': 'StatDelta',
  '2': [
    {'1': 'stat', '3': 1, '4': 1, '5': 9, '10': 'stat'},
    {'1': 'before', '3': 2, '4': 1, '5': 1, '10': 'before'},
    {'1': 'after', '3': 3, '4': 1, '5': 1, '10': 'after'},
    {'1': 'difference', '3': 4, '4': 1, '5': 1, '10': 'difference'},
  ],
};

/// Descriptor for `StatDelta`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List statDeltaDescriptor = $convert.base64Decode(
    'CglTdGF0RGVsdGESEgoEc3RhdBgBIAEoCVIEc3RhdBIWCgZiZWZvcmUYAiABKAFSBmJlZm9yZR'
    'IUCgVhZnRlchgDIAEoAVIFYWZ0ZXISHgoKZGlmZmVyZW5jZRgEIAEoAVIKZGlmZmVyZW5jZQ==');

@$core.Deprecated('Use buildComparisonDescriptor instead')
const BuildComparison$json = {
  '1': 'BuildComparison',
  '2': [
    {
      '1': 'before',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.CalculationResult',
      '10': 'before'
    },
    {
      '1': 'after',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.CalculationResult',
      '10': 'after'
    },
    {
      '1': 'deltas',
      '3': 3,
      '4': 3,
      '5': 11,
      '6': '.pob2.v1.StatDelta',
      '10': 'deltas'
    },
  ],
};

/// Descriptor for `BuildComparison`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List buildComparisonDescriptor = $convert.base64Decode(
    'Cg9CdWlsZENvbXBhcmlzb24SMgoGYmVmb3JlGAEgASgLMhoucG9iMi52MS5DYWxjdWxhdGlvbl'
    'Jlc3VsdFIGYmVmb3JlEjAKBWFmdGVyGAIgASgLMhoucG9iMi52MS5DYWxjdWxhdGlvblJlc3Vs'
    'dFIFYWZ0ZXISKgoGZGVsdGFzGAMgAygLMhIucG9iMi52MS5TdGF0RGVsdGFSBmRlbHRhcw==');
