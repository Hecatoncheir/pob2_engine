// This is a generated file - do not edit.
//
// Generated from pob2/v1/session.proto.

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

@$core.Deprecated('Use sessionSnapshotDescriptor instead')
const SessionSnapshot$json = {
  '1': 'SessionSnapshot',
  '2': [
    {'1': 'session_id', '3': 1, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'build_name', '3': 2, '4': 1, '5': 9, '10': 'buildName'},
    {'1': 'revision', '3': 3, '4': 1, '5': 4, '10': 'revision'},
  ],
};

/// Descriptor for `SessionSnapshot`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sessionSnapshotDescriptor = $convert.base64Decode(
    'Cg9TZXNzaW9uU25hcHNob3QSHQoKc2Vzc2lvbl9pZBgBIAEoCVIJc2Vzc2lvbklkEh0KCmJ1aW'
    'xkX25hbWUYAiABKAlSCWJ1aWxkTmFtZRIaCghyZXZpc2lvbhgDIAEoBFIIcmV2aXNpb24=');

@$core.Deprecated('Use getServerInfoRequestDescriptor instead')
const GetServerInfoRequest$json = {
  '1': 'GetServerInfoRequest',
};

/// Descriptor for `GetServerInfoRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getServerInfoRequestDescriptor =
    $convert.base64Decode('ChRHZXRTZXJ2ZXJJbmZvUmVxdWVzdA==');

@$core.Deprecated('Use getServerInfoResponseDescriptor instead')
const GetServerInfoResponse$json = {
  '1': 'GetServerInfoResponse',
  '2': [
    {'1': 'version', '3': 1, '4': 1, '5': 9, '10': 'version'},
    {'1': 'protocol_version', '3': 2, '4': 1, '5': 9, '10': 'protocolVersion'},
    {'1': 'capabilities', '3': 3, '4': 3, '5': 9, '10': 'capabilities'},
    {'1': 'state', '3': 4, '4': 1, '5': 9, '10': 'state'},
    {'1': 'instance_id', '3': 5, '4': 1, '5': 9, '10': 'instanceId'},
    {'1': 'mode', '3': 6, '4': 1, '5': 9, '10': 'mode'},
    {'1': 'client_count', '3': 7, '4': 1, '5': 13, '10': 'clientCount'},
    {'1': 'active_tasks', '3': 8, '4': 1, '5': 13, '10': 'activeTasks'},
    {'1': 'lease_timeout_ms', '3': 9, '4': 1, '5': 4, '10': 'leaseTimeoutMs'},
    {'1': 'storage_mode', '3': 10, '4': 1, '5': 9, '10': 'storageMode'},
    {'1': 'storage_path', '3': 11, '4': 1, '5': 9, '10': 'storagePath'},
  ],
};

/// Descriptor for `GetServerInfoResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getServerInfoResponseDescriptor = $convert.base64Decode(
    'ChVHZXRTZXJ2ZXJJbmZvUmVzcG9uc2USGAoHdmVyc2lvbhgBIAEoCVIHdmVyc2lvbhIpChBwcm'
    '90b2NvbF92ZXJzaW9uGAIgASgJUg9wcm90b2NvbFZlcnNpb24SIgoMY2FwYWJpbGl0aWVzGAMg'
    'AygJUgxjYXBhYmlsaXRpZXMSFAoFc3RhdGUYBCABKAlSBXN0YXRlEh8KC2luc3RhbmNlX2lkGA'
    'UgASgJUgppbnN0YW5jZUlkEhIKBG1vZGUYBiABKAlSBG1vZGUSIQoMY2xpZW50X2NvdW50GAcg'
    'ASgNUgtjbGllbnRDb3VudBIhCgxhY3RpdmVfdGFza3MYCCABKA1SC2FjdGl2ZVRhc2tzEigKEG'
    'xlYXNlX3RpbWVvdXRfbXMYCSABKARSDmxlYXNlVGltZW91dE1zEiEKDHN0b3JhZ2VfbW9kZRgK'
    'IAEoCVILc3RvcmFnZU1vZGUSIQoMc3RvcmFnZV9wYXRoGAsgASgJUgtzdG9yYWdlUGF0aA==');

@$core.Deprecated('Use createSessionRequestDescriptor instead')
const CreateSessionRequest$json = {
  '1': 'CreateSessionRequest',
  '2': [
    {'1': 'session_id', '3': 1, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'build_name', '3': 2, '4': 1, '5': 9, '10': 'buildName'},
    {'1': 'request_id', '3': 3, '4': 1, '5': 9, '10': 'requestId'},
  ],
};

/// Descriptor for `CreateSessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSessionRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVTZXNzaW9uUmVxdWVzdBIdCgpzZXNzaW9uX2lkGAEgASgJUglzZXNzaW9uSWQSHQ'
    'oKYnVpbGRfbmFtZRgCIAEoCVIJYnVpbGROYW1lEh0KCnJlcXVlc3RfaWQYAyABKAlSCXJlcXVl'
    'c3RJZA==');

@$core.Deprecated('Use createSessionResponseDescriptor instead')
const CreateSessionResponse$json = {
  '1': 'CreateSessionResponse',
  '2': [
    {
      '1': 'snapshot',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.SessionSnapshot',
      '10': 'snapshot'
    },
  ],
};

/// Descriptor for `CreateSessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSessionResponseDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVTZXNzaW9uUmVzcG9uc2USNAoIc25hcHNob3QYASABKAsyGC5wb2IyLnYxLlNlc3'
    'Npb25TbmFwc2hvdFIIc25hcHNob3Q=');

@$core.Deprecated('Use getSnapshotRequestDescriptor instead')
const GetSnapshotRequest$json = {
  '1': 'GetSnapshotRequest',
  '2': [
    {'1': 'session_id', '3': 1, '4': 1, '5': 9, '10': 'sessionId'},
  ],
};

/// Descriptor for `GetSnapshotRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSnapshotRequestDescriptor =
    $convert.base64Decode(
        'ChJHZXRTbmFwc2hvdFJlcXVlc3QSHQoKc2Vzc2lvbl9pZBgBIAEoCVIJc2Vzc2lvbklk');

@$core.Deprecated('Use getSnapshotResponseDescriptor instead')
const GetSnapshotResponse$json = {
  '1': 'GetSnapshotResponse',
  '2': [
    {
      '1': 'snapshot',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.SessionSnapshot',
      '10': 'snapshot'
    },
  ],
};

/// Descriptor for `GetSnapshotResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSnapshotResponseDescriptor = $convert.base64Decode(
    'ChNHZXRTbmFwc2hvdFJlc3BvbnNlEjQKCHNuYXBzaG90GAEgASgLMhgucG9iMi52MS5TZXNzaW'
    '9uU25hcHNob3RSCHNuYXBzaG90');

@$core.Deprecated('Use renameBuildRequestDescriptor instead')
const RenameBuildRequest$json = {
  '1': 'RenameBuildRequest',
  '2': [
    {'1': 'session_id', '3': 1, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'build_name', '3': 2, '4': 1, '5': 9, '10': 'buildName'},
    {
      '1': 'expected_revision',
      '3': 3,
      '4': 1,
      '5': 4,
      '10': 'expectedRevision'
    },
    {'1': 'request_id', '3': 4, '4': 1, '5': 9, '10': 'requestId'},
  ],
};

/// Descriptor for `RenameBuildRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List renameBuildRequestDescriptor = $convert.base64Decode(
    'ChJSZW5hbWVCdWlsZFJlcXVlc3QSHQoKc2Vzc2lvbl9pZBgBIAEoCVIJc2Vzc2lvbklkEh0KCm'
    'J1aWxkX25hbWUYAiABKAlSCWJ1aWxkTmFtZRIrChFleHBlY3RlZF9yZXZpc2lvbhgDIAEoBFIQ'
    'ZXhwZWN0ZWRSZXZpc2lvbhIdCgpyZXF1ZXN0X2lkGAQgASgJUglyZXF1ZXN0SWQ=');

@$core.Deprecated('Use renameBuildResponseDescriptor instead')
const RenameBuildResponse$json = {
  '1': 'RenameBuildResponse',
  '2': [
    {
      '1': 'snapshot',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.SessionSnapshot',
      '10': 'snapshot'
    },
  ],
};

/// Descriptor for `RenameBuildResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List renameBuildResponseDescriptor = $convert.base64Decode(
    'ChNSZW5hbWVCdWlsZFJlc3BvbnNlEjQKCHNuYXBzaG90GAEgASgLMhgucG9iMi52MS5TZXNzaW'
    '9uU25hcHNob3RSCHNuYXBzaG90');

@$core.Deprecated('Use watchSessionRequestDescriptor instead')
const WatchSessionRequest$json = {
  '1': 'WatchSessionRequest',
  '2': [
    {'1': 'session_id', '3': 1, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'after_revision', '3': 2, '4': 1, '5': 4, '10': 'afterRevision'},
  ],
};

/// Descriptor for `WatchSessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List watchSessionRequestDescriptor = $convert.base64Decode(
    'ChNXYXRjaFNlc3Npb25SZXF1ZXN0Eh0KCnNlc3Npb25faWQYASABKAlSCXNlc3Npb25JZBIlCg'
    '5hZnRlcl9yZXZpc2lvbhgCIAEoBFINYWZ0ZXJSZXZpc2lvbg==');

@$core.Deprecated('Use watchSessionResponseDescriptor instead')
const WatchSessionResponse$json = {
  '1': 'WatchSessionResponse',
  '2': [
    {
      '1': 'snapshot',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.pob2.v1.SessionSnapshot',
      '10': 'snapshot'
    },
    {'1': 'kind', '3': 2, '4': 1, '5': 9, '10': 'kind'},
    {'1': 'request_id', '3': 3, '4': 1, '5': 9, '10': 'requestId'},
    {'1': 'changed_fields', '3': 4, '4': 3, '5': 9, '10': 'changedFields'},
    {
      '1': 'previous_build_name',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'previousBuildName'
    },
  ],
};

/// Descriptor for `WatchSessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List watchSessionResponseDescriptor = $convert.base64Decode(
    'ChRXYXRjaFNlc3Npb25SZXNwb25zZRI0CghzbmFwc2hvdBgBIAEoCzIYLnBvYjIudjEuU2Vzc2'
    'lvblNuYXBzaG90UghzbmFwc2hvdBISCgRraW5kGAIgASgJUgRraW5kEh0KCnJlcXVlc3RfaWQY'
    'AyABKAlSCXJlcXVlc3RJZBIlCg5jaGFuZ2VkX2ZpZWxkcxgEIAMoCVINY2hhbmdlZEZpZWxkcx'
    'IuChNwcmV2aW91c19idWlsZF9uYW1lGAUgASgJUhFwcmV2aW91c0J1aWxkTmFtZQ==');

@$core.Deprecated('Use shutdownRequestDescriptor instead')
const ShutdownRequest$json = {
  '1': 'ShutdownRequest',
};

/// Descriptor for `ShutdownRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List shutdownRequestDescriptor =
    $convert.base64Decode('Cg9TaHV0ZG93blJlcXVlc3Q=');

@$core.Deprecated('Use shutdownResponseDescriptor instead')
const ShutdownResponse$json = {
  '1': 'ShutdownResponse',
};

/// Descriptor for `ShutdownResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List shutdownResponseDescriptor =
    $convert.base64Decode('ChBTaHV0ZG93blJlc3BvbnNl');

@$core.Deprecated('Use registerClientRequestDescriptor instead')
const RegisterClientRequest$json = {
  '1': 'RegisterClientRequest',
  '2': [
    {'1': 'client_id', '3': 1, '4': 1, '5': 9, '10': 'clientId'},
    {'1': 'kind', '3': 2, '4': 1, '5': 9, '10': 'kind'},
  ],
};

/// Descriptor for `RegisterClientRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List registerClientRequestDescriptor = $convert.base64Decode(
    'ChVSZWdpc3RlckNsaWVudFJlcXVlc3QSGwoJY2xpZW50X2lkGAEgASgJUghjbGllbnRJZBISCg'
    'RraW5kGAIgASgJUgRraW5k');

@$core.Deprecated('Use registerClientResponseDescriptor instead')
const RegisterClientResponse$json = {
  '1': 'RegisterClientResponse',
  '2': [
    {'1': 'lease_id', '3': 1, '4': 1, '5': 9, '10': 'leaseId'},
    {'1': 'lease_timeout_ms', '3': 2, '4': 1, '5': 4, '10': 'leaseTimeoutMs'},
  ],
};

/// Descriptor for `RegisterClientResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List registerClientResponseDescriptor =
    $convert.base64Decode(
        'ChZSZWdpc3RlckNsaWVudFJlc3BvbnNlEhkKCGxlYXNlX2lkGAEgASgJUgdsZWFzZUlkEigKEG'
        'xlYXNlX3RpbWVvdXRfbXMYAiABKARSDmxlYXNlVGltZW91dE1z');

@$core.Deprecated('Use heartbeatRequestDescriptor instead')
const HeartbeatRequest$json = {
  '1': 'HeartbeatRequest',
  '2': [
    {'1': 'lease_id', '3': 1, '4': 1, '5': 9, '10': 'leaseId'},
  ],
};

/// Descriptor for `HeartbeatRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List heartbeatRequestDescriptor = $convert.base64Decode(
    'ChBIZWFydGJlYXRSZXF1ZXN0EhkKCGxlYXNlX2lkGAEgASgJUgdsZWFzZUlk');

@$core.Deprecated('Use heartbeatResponseDescriptor instead')
const HeartbeatResponse$json = {
  '1': 'HeartbeatResponse',
};

/// Descriptor for `HeartbeatResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List heartbeatResponseDescriptor =
    $convert.base64Decode('ChFIZWFydGJlYXRSZXNwb25zZQ==');

@$core.Deprecated('Use releaseClientRequestDescriptor instead')
const ReleaseClientRequest$json = {
  '1': 'ReleaseClientRequest',
  '2': [
    {'1': 'lease_id', '3': 1, '4': 1, '5': 9, '10': 'leaseId'},
  ],
};

/// Descriptor for `ReleaseClientRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List releaseClientRequestDescriptor =
    $convert.base64Decode(
        'ChRSZWxlYXNlQ2xpZW50UmVxdWVzdBIZCghsZWFzZV9pZBgBIAEoCVIHbGVhc2VJZA==');

@$core.Deprecated('Use releaseClientResponseDescriptor instead')
const ReleaseClientResponse$json = {
  '1': 'ReleaseClientResponse',
};

/// Descriptor for `ReleaseClientResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List releaseClientResponseDescriptor =
    $convert.base64Decode('ChVSZWxlYXNlQ2xpZW50UmVzcG9uc2U=');
