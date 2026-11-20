// This is a generated file - do not edit.
//
// Generated from pob2/v1/session.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class SessionSnapshot extends $pb.GeneratedMessage {
  factory SessionSnapshot({
    $core.String? sessionId,
    $core.String? buildName,
    $fixnum.Int64? revision,
  }) {
    final result = SessionSnapshot._();
    if (sessionId != null) result.sessionId = sessionId;
    if (buildName != null) result.buildName = buildName;
    if (revision != null) result.revision = revision;
    return result;
  }

  SessionSnapshot._();

  factory SessionSnapshot.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SessionSnapshot()..mergeFromBuffer(data, registry);
  factory SessionSnapshot.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SessionSnapshot()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SessionSnapshot',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: SessionSnapshot.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'sessionId')
    ..aOS(2, _omitFieldNames ? '' : 'buildName')
    ..a<$fixnum.Int64>(
        3, _omitFieldNames ? '' : 'revision', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SessionSnapshot clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SessionSnapshot copyWith(void Function(SessionSnapshot) updates) =>
      super.copyWith((message) => updates(message as SessionSnapshot))
          as SessionSnapshot;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SessionSnapshot() / SessionSnapshot.new instead')
  static SessionSnapshot create() => SessionSnapshot._();
  static $pb.GeneratedMessage $_createMessage() => SessionSnapshot._();
  @$core.override
  SessionSnapshot createEmptyInstance() => SessionSnapshot._();
  @$core.pragma('dart2js:noInline')
  static SessionSnapshot getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SessionSnapshot>(
          SessionSnapshot.$_createMessage);
  static SessionSnapshot? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sessionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set sessionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSessionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSessionId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get buildName => $_getSZ(1);
  @$pb.TagNumber(2)
  set buildName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBuildName() => $_has(1);
  @$pb.TagNumber(2)
  void clearBuildName() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get revision => $_getI64(2);
  @$pb.TagNumber(3)
  set revision($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRevision() => $_has(2);
  @$pb.TagNumber(3)
  void clearRevision() => $_clearField(3);
}

class GetServerInfoRequest extends $pb.GeneratedMessage {
  factory GetServerInfoRequest() => GetServerInfoRequest._();

  GetServerInfoRequest._();

  factory GetServerInfoRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetServerInfoRequest()..mergeFromBuffer(data, registry);
  factory GetServerInfoRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetServerInfoRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetServerInfoRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: GetServerInfoRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetServerInfoRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetServerInfoRequest copyWith(void Function(GetServerInfoRequest) updates) =>
      super.copyWith((message) => updates(message as GetServerInfoRequest))
          as GetServerInfoRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetServerInfoRequest() / GetServerInfoRequest.new instead')
  static GetServerInfoRequest create() => GetServerInfoRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetServerInfoRequest._();
  @$core.override
  GetServerInfoRequest createEmptyInstance() => GetServerInfoRequest._();
  @$core.pragma('dart2js:noInline')
  static GetServerInfoRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetServerInfoRequest>(
          GetServerInfoRequest.$_createMessage);
  static GetServerInfoRequest? _defaultInstance;
}

class GetServerInfoResponse extends $pb.GeneratedMessage {
  factory GetServerInfoResponse({
    $core.String? version,
    $core.String? protocolVersion,
    $core.Iterable<$core.String>? capabilities,
    $core.String? state,
    $core.String? instanceId,
    $core.String? mode,
    $core.int? clientCount,
    $core.int? activeTasks,
    $fixnum.Int64? leaseTimeoutMs,
    $core.String? storageMode,
    $core.String? storagePath,
  }) {
    final result = GetServerInfoResponse._();
    if (version != null) result.version = version;
    if (protocolVersion != null) result.protocolVersion = protocolVersion;
    if (capabilities != null) result.capabilities.addAll(capabilities);
    if (state != null) result.state = state;
    if (instanceId != null) result.instanceId = instanceId;
    if (mode != null) result.mode = mode;
    if (clientCount != null) result.clientCount = clientCount;
    if (activeTasks != null) result.activeTasks = activeTasks;
    if (leaseTimeoutMs != null) result.leaseTimeoutMs = leaseTimeoutMs;
    if (storageMode != null) result.storageMode = storageMode;
    if (storagePath != null) result.storagePath = storagePath;
    return result;
  }

  GetServerInfoResponse._();

  factory GetServerInfoResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetServerInfoResponse()..mergeFromBuffer(data, registry);
  factory GetServerInfoResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetServerInfoResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetServerInfoResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: GetServerInfoResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'version')
    ..aOS(2, _omitFieldNames ? '' : 'protocolVersion')
    ..pPS(3, _omitFieldNames ? '' : 'capabilities')
    ..aOS(4, _omitFieldNames ? '' : 'state')
    ..aOS(5, _omitFieldNames ? '' : 'instanceId')
    ..aOS(6, _omitFieldNames ? '' : 'mode')
    ..aI(7, _omitFieldNames ? '' : 'clientCount',
        fieldType: $pb.PbFieldType.OU3)
    ..aI(8, _omitFieldNames ? '' : 'activeTasks',
        fieldType: $pb.PbFieldType.OU3)
    ..a<$fixnum.Int64>(
        9, _omitFieldNames ? '' : 'leaseTimeoutMs', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(10, _omitFieldNames ? '' : 'storageMode')
    ..aOS(11, _omitFieldNames ? '' : 'storagePath')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetServerInfoResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetServerInfoResponse copyWith(
          void Function(GetServerInfoResponse) updates) =>
      super.copyWith((message) => updates(message as GetServerInfoResponse))
          as GetServerInfoResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetServerInfoResponse() / GetServerInfoResponse.new instead')
  static GetServerInfoResponse create() => GetServerInfoResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetServerInfoResponse._();
  @$core.override
  GetServerInfoResponse createEmptyInstance() => GetServerInfoResponse._();
  @$core.pragma('dart2js:noInline')
  static GetServerInfoResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetServerInfoResponse>(
          GetServerInfoResponse.$_createMessage);
  static GetServerInfoResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get version => $_getSZ(0);
  @$pb.TagNumber(1)
  set version($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasVersion() => $_has(0);
  @$pb.TagNumber(1)
  void clearVersion() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get protocolVersion => $_getSZ(1);
  @$pb.TagNumber(2)
  set protocolVersion($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProtocolVersion() => $_has(1);
  @$pb.TagNumber(2)
  void clearProtocolVersion() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get capabilities => $_getList(2);

  @$pb.TagNumber(4)
  $core.String get state => $_getSZ(3);
  @$pb.TagNumber(4)
  set state($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasState() => $_has(3);
  @$pb.TagNumber(4)
  void clearState() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get instanceId => $_getSZ(4);
  @$pb.TagNumber(5)
  set instanceId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasInstanceId() => $_has(4);
  @$pb.TagNumber(5)
  void clearInstanceId() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get mode => $_getSZ(5);
  @$pb.TagNumber(6)
  set mode($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasMode() => $_has(5);
  @$pb.TagNumber(6)
  void clearMode() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.int get clientCount => $_getIZ(6);
  @$pb.TagNumber(7)
  set clientCount($core.int value) => $_setUnsignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasClientCount() => $_has(6);
  @$pb.TagNumber(7)
  void clearClientCount() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.int get activeTasks => $_getIZ(7);
  @$pb.TagNumber(8)
  set activeTasks($core.int value) => $_setUnsignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasActiveTasks() => $_has(7);
  @$pb.TagNumber(8)
  void clearActiveTasks() => $_clearField(8);

  @$pb.TagNumber(9)
  $fixnum.Int64 get leaseTimeoutMs => $_getI64(8);
  @$pb.TagNumber(9)
  set leaseTimeoutMs($fixnum.Int64 value) => $_setInt64(8, value);
  @$pb.TagNumber(9)
  $core.bool hasLeaseTimeoutMs() => $_has(8);
  @$pb.TagNumber(9)
  void clearLeaseTimeoutMs() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.String get storageMode => $_getSZ(9);
  @$pb.TagNumber(10)
  set storageMode($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasStorageMode() => $_has(9);
  @$pb.TagNumber(10)
  void clearStorageMode() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.String get storagePath => $_getSZ(10);
  @$pb.TagNumber(11)
  set storagePath($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasStoragePath() => $_has(10);
  @$pb.TagNumber(11)
  void clearStoragePath() => $_clearField(11);
}

/// request_id повторяется только для точного повтора той же команды.
class CreateSessionRequest extends $pb.GeneratedMessage {
  factory CreateSessionRequest({
    $core.String? sessionId,
    $core.String? buildName,
    $core.String? requestId,
  }) {
    final result = CreateSessionRequest._();
    if (sessionId != null) result.sessionId = sessionId;
    if (buildName != null) result.buildName = buildName;
    if (requestId != null) result.requestId = requestId;
    return result;
  }

  CreateSessionRequest._();

  factory CreateSessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSessionRequest()..mergeFromBuffer(data, registry);
  factory CreateSessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSessionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateSessionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: CreateSessionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'sessionId')
    ..aOS(2, _omitFieldNames ? '' : 'buildName')
    ..aOS(3, _omitFieldNames ? '' : 'requestId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSessionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSessionRequest copyWith(void Function(CreateSessionRequest) updates) =>
      super.copyWith((message) => updates(message as CreateSessionRequest))
          as CreateSessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateSessionRequest() / CreateSessionRequest.new instead')
  static CreateSessionRequest create() => CreateSessionRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateSessionRequest._();
  @$core.override
  CreateSessionRequest createEmptyInstance() => CreateSessionRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateSessionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateSessionRequest>(
          CreateSessionRequest.$_createMessage);
  static CreateSessionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sessionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set sessionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSessionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSessionId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get buildName => $_getSZ(1);
  @$pb.TagNumber(2)
  set buildName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBuildName() => $_has(1);
  @$pb.TagNumber(2)
  void clearBuildName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get requestId => $_getSZ(2);
  @$pb.TagNumber(3)
  set requestId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRequestId() => $_has(2);
  @$pb.TagNumber(3)
  void clearRequestId() => $_clearField(3);
}

class CreateSessionResponse extends $pb.GeneratedMessage {
  factory CreateSessionResponse({
    SessionSnapshot? snapshot,
  }) {
    final result = CreateSessionResponse._();
    if (snapshot != null) result.snapshot = snapshot;
    return result;
  }

  CreateSessionResponse._();

  factory CreateSessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSessionResponse()..mergeFromBuffer(data, registry);
  factory CreateSessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSessionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateSessionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: CreateSessionResponse.$_createMessage)
    ..aOM<SessionSnapshot>(1, _omitFieldNames ? '' : 'snapshot',
        subBuilder: SessionSnapshot.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSessionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSessionResponse copyWith(
          void Function(CreateSessionResponse) updates) =>
      super.copyWith((message) => updates(message as CreateSessionResponse))
          as CreateSessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateSessionResponse() / CreateSessionResponse.new instead')
  static CreateSessionResponse create() => CreateSessionResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateSessionResponse._();
  @$core.override
  CreateSessionResponse createEmptyInstance() => CreateSessionResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateSessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateSessionResponse>(
          CreateSessionResponse.$_createMessage);
  static CreateSessionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SessionSnapshot get snapshot => $_getN(0);
  @$pb.TagNumber(1)
  set snapshot(SessionSnapshot value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSnapshot() => $_has(0);
  @$pb.TagNumber(1)
  void clearSnapshot() => $_clearField(1);
  @$pb.TagNumber(1)
  SessionSnapshot ensureSnapshot() => $_ensure(0);
}

class GetSnapshotRequest extends $pb.GeneratedMessage {
  factory GetSnapshotRequest({
    $core.String? sessionId,
  }) {
    final result = GetSnapshotRequest._();
    if (sessionId != null) result.sessionId = sessionId;
    return result;
  }

  GetSnapshotRequest._();

  factory GetSnapshotRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSnapshotRequest()..mergeFromBuffer(data, registry);
  factory GetSnapshotRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSnapshotRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetSnapshotRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: GetSnapshotRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'sessionId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSnapshotRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSnapshotRequest copyWith(void Function(GetSnapshotRequest) updates) =>
      super.copyWith((message) => updates(message as GetSnapshotRequest))
          as GetSnapshotRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetSnapshotRequest() / GetSnapshotRequest.new instead')
  static GetSnapshotRequest create() => GetSnapshotRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetSnapshotRequest._();
  @$core.override
  GetSnapshotRequest createEmptyInstance() => GetSnapshotRequest._();
  @$core.pragma('dart2js:noInline')
  static GetSnapshotRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetSnapshotRequest>(
          GetSnapshotRequest.$_createMessage);
  static GetSnapshotRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sessionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set sessionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSessionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSessionId() => $_clearField(1);
}

class GetSnapshotResponse extends $pb.GeneratedMessage {
  factory GetSnapshotResponse({
    SessionSnapshot? snapshot,
  }) {
    final result = GetSnapshotResponse._();
    if (snapshot != null) result.snapshot = snapshot;
    return result;
  }

  GetSnapshotResponse._();

  factory GetSnapshotResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSnapshotResponse()..mergeFromBuffer(data, registry);
  factory GetSnapshotResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSnapshotResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetSnapshotResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: GetSnapshotResponse.$_createMessage)
    ..aOM<SessionSnapshot>(1, _omitFieldNames ? '' : 'snapshot',
        subBuilder: SessionSnapshot.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSnapshotResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSnapshotResponse copyWith(void Function(GetSnapshotResponse) updates) =>
      super.copyWith((message) => updates(message as GetSnapshotResponse))
          as GetSnapshotResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use GetSnapshotResponse() / GetSnapshotResponse.new instead')
  static GetSnapshotResponse create() => GetSnapshotResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetSnapshotResponse._();
  @$core.override
  GetSnapshotResponse createEmptyInstance() => GetSnapshotResponse._();
  @$core.pragma('dart2js:noInline')
  static GetSnapshotResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetSnapshotResponse>(
          GetSnapshotResponse.$_createMessage);
  static GetSnapshotResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SessionSnapshot get snapshot => $_getN(0);
  @$pb.TagNumber(1)
  set snapshot(SessionSnapshot value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSnapshot() => $_has(0);
  @$pb.TagNumber(1)
  void clearSnapshot() => $_clearField(1);
  @$pb.TagNumber(1)
  SessionSnapshot ensureSnapshot() => $_ensure(0);
}

/// Точный повтор возвращает исходный результат даже после последующих правок.
class RenameBuildRequest extends $pb.GeneratedMessage {
  factory RenameBuildRequest({
    $core.String? sessionId,
    $core.String? buildName,
    $fixnum.Int64? expectedRevision,
    $core.String? requestId,
  }) {
    final result = RenameBuildRequest._();
    if (sessionId != null) result.sessionId = sessionId;
    if (buildName != null) result.buildName = buildName;
    if (expectedRevision != null) result.expectedRevision = expectedRevision;
    if (requestId != null) result.requestId = requestId;
    return result;
  }

  RenameBuildRequest._();

  factory RenameBuildRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RenameBuildRequest()..mergeFromBuffer(data, registry);
  factory RenameBuildRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RenameBuildRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RenameBuildRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: RenameBuildRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'sessionId')
    ..aOS(2, _omitFieldNames ? '' : 'buildName')
    ..a<$fixnum.Int64>(
        3, _omitFieldNames ? '' : 'expectedRevision', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(4, _omitFieldNames ? '' : 'requestId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RenameBuildRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RenameBuildRequest copyWith(void Function(RenameBuildRequest) updates) =>
      super.copyWith((message) => updates(message as RenameBuildRequest))
          as RenameBuildRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RenameBuildRequest() / RenameBuildRequest.new instead')
  static RenameBuildRequest create() => RenameBuildRequest._();
  static $pb.GeneratedMessage $_createMessage() => RenameBuildRequest._();
  @$core.override
  RenameBuildRequest createEmptyInstance() => RenameBuildRequest._();
  @$core.pragma('dart2js:noInline')
  static RenameBuildRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RenameBuildRequest>(
          RenameBuildRequest.$_createMessage);
  static RenameBuildRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sessionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set sessionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSessionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSessionId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get buildName => $_getSZ(1);
  @$pb.TagNumber(2)
  set buildName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBuildName() => $_has(1);
  @$pb.TagNumber(2)
  void clearBuildName() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get expectedRevision => $_getI64(2);
  @$pb.TagNumber(3)
  set expectedRevision($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasExpectedRevision() => $_has(2);
  @$pb.TagNumber(3)
  void clearExpectedRevision() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get requestId => $_getSZ(3);
  @$pb.TagNumber(4)
  set requestId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRequestId() => $_has(3);
  @$pb.TagNumber(4)
  void clearRequestId() => $_clearField(4);
}

class RenameBuildResponse extends $pb.GeneratedMessage {
  factory RenameBuildResponse({
    SessionSnapshot? snapshot,
  }) {
    final result = RenameBuildResponse._();
    if (snapshot != null) result.snapshot = snapshot;
    return result;
  }

  RenameBuildResponse._();

  factory RenameBuildResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RenameBuildResponse()..mergeFromBuffer(data, registry);
  factory RenameBuildResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RenameBuildResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RenameBuildResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: RenameBuildResponse.$_createMessage)
    ..aOM<SessionSnapshot>(1, _omitFieldNames ? '' : 'snapshot',
        subBuilder: SessionSnapshot.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RenameBuildResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RenameBuildResponse copyWith(void Function(RenameBuildResponse) updates) =>
      super.copyWith((message) => updates(message as RenameBuildResponse))
          as RenameBuildResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use RenameBuildResponse() / RenameBuildResponse.new instead')
  static RenameBuildResponse create() => RenameBuildResponse._();
  static $pb.GeneratedMessage $_createMessage() => RenameBuildResponse._();
  @$core.override
  RenameBuildResponse createEmptyInstance() => RenameBuildResponse._();
  @$core.pragma('dart2js:noInline')
  static RenameBuildResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RenameBuildResponse>(
          RenameBuildResponse.$_createMessage);
  static RenameBuildResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SessionSnapshot get snapshot => $_getN(0);
  @$pb.TagNumber(1)
  set snapshot(SessionSnapshot value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSnapshot() => $_has(0);
  @$pb.TagNumber(1)
  void clearSnapshot() => $_clearField(1);
  @$pb.TagNumber(1)
  SessionSnapshot ensureSnapshot() => $_ensure(0);
}

/// Ноль запрашивает начальный снимок; иначе возвращаются все последующие правки.
class WatchSessionRequest extends $pb.GeneratedMessage {
  factory WatchSessionRequest({
    $core.String? sessionId,
    $fixnum.Int64? afterRevision,
  }) {
    final result = WatchSessionRequest._();
    if (sessionId != null) result.sessionId = sessionId;
    if (afterRevision != null) result.afterRevision = afterRevision;
    return result;
  }

  WatchSessionRequest._();

  factory WatchSessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WatchSessionRequest()..mergeFromBuffer(data, registry);
  factory WatchSessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WatchSessionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WatchSessionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: WatchSessionRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'sessionId')
    ..a<$fixnum.Int64>(
        2, _omitFieldNames ? '' : 'afterRevision', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WatchSessionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WatchSessionRequest copyWith(void Function(WatchSessionRequest) updates) =>
      super.copyWith((message) => updates(message as WatchSessionRequest))
          as WatchSessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use WatchSessionRequest() / WatchSessionRequest.new instead')
  static WatchSessionRequest create() => WatchSessionRequest._();
  static $pb.GeneratedMessage $_createMessage() => WatchSessionRequest._();
  @$core.override
  WatchSessionRequest createEmptyInstance() => WatchSessionRequest._();
  @$core.pragma('dart2js:noInline')
  static WatchSessionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WatchSessionRequest>(
          WatchSessionRequest.$_createMessage);
  static WatchSessionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get sessionId => $_getSZ(0);
  @$pb.TagNumber(1)
  set sessionId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSessionId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSessionId() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get afterRevision => $_getI64(1);
  @$pb.TagNumber(2)
  set afterRevision($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAfterRevision() => $_has(1);
  @$pb.TagNumber(2)
  void clearAfterRevision() => $_clearField(2);
}

class WatchSessionResponse extends $pb.GeneratedMessage {
  factory WatchSessionResponse({
    SessionSnapshot? snapshot,
    $core.String? kind,
    $core.String? requestId,
    $core.Iterable<$core.String>? changedFields,
    $core.String? previousBuildName,
  }) {
    final result = WatchSessionResponse._();
    if (snapshot != null) result.snapshot = snapshot;
    if (kind != null) result.kind = kind;
    if (requestId != null) result.requestId = requestId;
    if (changedFields != null) result.changedFields.addAll(changedFields);
    if (previousBuildName != null) result.previousBuildName = previousBuildName;
    return result;
  }

  WatchSessionResponse._();

  factory WatchSessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WatchSessionResponse()..mergeFromBuffer(data, registry);
  factory WatchSessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WatchSessionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WatchSessionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: WatchSessionResponse.$_createMessage)
    ..aOM<SessionSnapshot>(1, _omitFieldNames ? '' : 'snapshot',
        subBuilder: SessionSnapshot.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'kind')
    ..aOS(3, _omitFieldNames ? '' : 'requestId')
    ..pPS(4, _omitFieldNames ? '' : 'changedFields')
    ..aOS(5, _omitFieldNames ? '' : 'previousBuildName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WatchSessionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WatchSessionResponse copyWith(void Function(WatchSessionResponse) updates) =>
      super.copyWith((message) => updates(message as WatchSessionResponse))
          as WatchSessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use WatchSessionResponse() / WatchSessionResponse.new instead')
  static WatchSessionResponse create() => WatchSessionResponse._();
  static $pb.GeneratedMessage $_createMessage() => WatchSessionResponse._();
  @$core.override
  WatchSessionResponse createEmptyInstance() => WatchSessionResponse._();
  @$core.pragma('dart2js:noInline')
  static WatchSessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WatchSessionResponse>(
          WatchSessionResponse.$_createMessage);
  static WatchSessionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SessionSnapshot get snapshot => $_getN(0);
  @$pb.TagNumber(1)
  set snapshot(SessionSnapshot value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSnapshot() => $_has(0);
  @$pb.TagNumber(1)
  void clearSnapshot() => $_clearField(1);
  @$pb.TagNumber(1)
  SessionSnapshot ensureSnapshot() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.String get kind => $_getSZ(1);
  @$pb.TagNumber(2)
  set kind($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasKind() => $_has(1);
  @$pb.TagNumber(2)
  void clearKind() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get requestId => $_getSZ(2);
  @$pb.TagNumber(3)
  set requestId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRequestId() => $_has(2);
  @$pb.TagNumber(3)
  void clearRequestId() => $_clearField(3);

  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get changedFields => $_getList(3);

  @$pb.TagNumber(5)
  $core.String get previousBuildName => $_getSZ(4);
  @$pb.TagNumber(5)
  set previousBuildName($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPreviousBuildName() => $_has(4);
  @$pb.TagNumber(5)
  void clearPreviousBuildName() => $_clearField(5);
}

class ShutdownRequest extends $pb.GeneratedMessage {
  factory ShutdownRequest() => ShutdownRequest._();

  ShutdownRequest._();

  factory ShutdownRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ShutdownRequest()..mergeFromBuffer(data, registry);
  factory ShutdownRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ShutdownRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ShutdownRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: ShutdownRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ShutdownRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ShutdownRequest copyWith(void Function(ShutdownRequest) updates) =>
      super.copyWith((message) => updates(message as ShutdownRequest))
          as ShutdownRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ShutdownRequest() / ShutdownRequest.new instead')
  static ShutdownRequest create() => ShutdownRequest._();
  static $pb.GeneratedMessage $_createMessage() => ShutdownRequest._();
  @$core.override
  ShutdownRequest createEmptyInstance() => ShutdownRequest._();
  @$core.pragma('dart2js:noInline')
  static ShutdownRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ShutdownRequest>(
          ShutdownRequest.$_createMessage);
  static ShutdownRequest? _defaultInstance;
}

class ShutdownResponse extends $pb.GeneratedMessage {
  factory ShutdownResponse() => ShutdownResponse._();

  ShutdownResponse._();

  factory ShutdownResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ShutdownResponse()..mergeFromBuffer(data, registry);
  factory ShutdownResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ShutdownResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ShutdownResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: ShutdownResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ShutdownResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ShutdownResponse copyWith(void Function(ShutdownResponse) updates) =>
      super.copyWith((message) => updates(message as ShutdownResponse))
          as ShutdownResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ShutdownResponse() / ShutdownResponse.new instead')
  static ShutdownResponse create() => ShutdownResponse._();
  static $pb.GeneratedMessage $_createMessage() => ShutdownResponse._();
  @$core.override
  ShutdownResponse createEmptyInstance() => ShutdownResponse._();
  @$core.pragma('dart2js:noInline')
  static ShutdownResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ShutdownResponse>(
          ShutdownResponse.$_createMessage);
  static ShutdownResponse? _defaultInstance;
}

/// client_id обозначает приложение, а не TCP-соединение или отдельный RPC.
class RegisterClientRequest extends $pb.GeneratedMessage {
  factory RegisterClientRequest({
    $core.String? clientId,
    $core.String? kind,
  }) {
    final result = RegisterClientRequest._();
    if (clientId != null) result.clientId = clientId;
    if (kind != null) result.kind = kind;
    return result;
  }

  RegisterClientRequest._();

  factory RegisterClientRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RegisterClientRequest()..mergeFromBuffer(data, registry);
  factory RegisterClientRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RegisterClientRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RegisterClientRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: RegisterClientRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..aOS(2, _omitFieldNames ? '' : 'kind')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterClientRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterClientRequest copyWith(
          void Function(RegisterClientRequest) updates) =>
      super.copyWith((message) => updates(message as RegisterClientRequest))
          as RegisterClientRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RegisterClientRequest() / RegisterClientRequest.new instead')
  static RegisterClientRequest create() => RegisterClientRequest._();
  static $pb.GeneratedMessage $_createMessage() => RegisterClientRequest._();
  @$core.override
  RegisterClientRequest createEmptyInstance() => RegisterClientRequest._();
  @$core.pragma('dart2js:noInline')
  static RegisterClientRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RegisterClientRequest>(
          RegisterClientRequest.$_createMessage);
  static RegisterClientRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get kind => $_getSZ(1);
  @$pb.TagNumber(2)
  set kind($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasKind() => $_has(1);
  @$pb.TagNumber(2)
  void clearKind() => $_clearField(2);
}

class RegisterClientResponse extends $pb.GeneratedMessage {
  factory RegisterClientResponse({
    $core.String? leaseId,
    $fixnum.Int64? leaseTimeoutMs,
  }) {
    final result = RegisterClientResponse._();
    if (leaseId != null) result.leaseId = leaseId;
    if (leaseTimeoutMs != null) result.leaseTimeoutMs = leaseTimeoutMs;
    return result;
  }

  RegisterClientResponse._();

  factory RegisterClientResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RegisterClientResponse()..mergeFromBuffer(data, registry);
  factory RegisterClientResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RegisterClientResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RegisterClientResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: RegisterClientResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'leaseId')
    ..a<$fixnum.Int64>(
        2, _omitFieldNames ? '' : 'leaseTimeoutMs', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterClientResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterClientResponse copyWith(
          void Function(RegisterClientResponse) updates) =>
      super.copyWith((message) => updates(message as RegisterClientResponse))
          as RegisterClientResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RegisterClientResponse() / RegisterClientResponse.new instead')
  static RegisterClientResponse create() => RegisterClientResponse._();
  static $pb.GeneratedMessage $_createMessage() => RegisterClientResponse._();
  @$core.override
  RegisterClientResponse createEmptyInstance() => RegisterClientResponse._();
  @$core.pragma('dart2js:noInline')
  static RegisterClientResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RegisterClientResponse>(
          RegisterClientResponse.$_createMessage);
  static RegisterClientResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get leaseId => $_getSZ(0);
  @$pb.TagNumber(1)
  set leaseId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasLeaseId() => $_has(0);
  @$pb.TagNumber(1)
  void clearLeaseId() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get leaseTimeoutMs => $_getI64(1);
  @$pb.TagNumber(2)
  set leaseTimeoutMs($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLeaseTimeoutMs() => $_has(1);
  @$pb.TagNumber(2)
  void clearLeaseTimeoutMs() => $_clearField(2);
}

class HeartbeatRequest extends $pb.GeneratedMessage {
  factory HeartbeatRequest({
    $core.String? leaseId,
  }) {
    final result = HeartbeatRequest._();
    if (leaseId != null) result.leaseId = leaseId;
    return result;
  }

  HeartbeatRequest._();

  factory HeartbeatRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HeartbeatRequest()..mergeFromBuffer(data, registry);
  factory HeartbeatRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HeartbeatRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'HeartbeatRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: HeartbeatRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'leaseId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HeartbeatRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HeartbeatRequest copyWith(void Function(HeartbeatRequest) updates) =>
      super.copyWith((message) => updates(message as HeartbeatRequest))
          as HeartbeatRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use HeartbeatRequest() / HeartbeatRequest.new instead')
  static HeartbeatRequest create() => HeartbeatRequest._();
  static $pb.GeneratedMessage $_createMessage() => HeartbeatRequest._();
  @$core.override
  HeartbeatRequest createEmptyInstance() => HeartbeatRequest._();
  @$core.pragma('dart2js:noInline')
  static HeartbeatRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<HeartbeatRequest>(
          HeartbeatRequest.$_createMessage);
  static HeartbeatRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get leaseId => $_getSZ(0);
  @$pb.TagNumber(1)
  set leaseId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasLeaseId() => $_has(0);
  @$pb.TagNumber(1)
  void clearLeaseId() => $_clearField(1);
}

class HeartbeatResponse extends $pb.GeneratedMessage {
  factory HeartbeatResponse() => HeartbeatResponse._();

  HeartbeatResponse._();

  factory HeartbeatResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HeartbeatResponse()..mergeFromBuffer(data, registry);
  factory HeartbeatResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HeartbeatResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'HeartbeatResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: HeartbeatResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HeartbeatResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HeartbeatResponse copyWith(void Function(HeartbeatResponse) updates) =>
      super.copyWith((message) => updates(message as HeartbeatResponse))
          as HeartbeatResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use HeartbeatResponse() / HeartbeatResponse.new instead')
  static HeartbeatResponse create() => HeartbeatResponse._();
  static $pb.GeneratedMessage $_createMessage() => HeartbeatResponse._();
  @$core.override
  HeartbeatResponse createEmptyInstance() => HeartbeatResponse._();
  @$core.pragma('dart2js:noInline')
  static HeartbeatResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<HeartbeatResponse>(
          HeartbeatResponse.$_createMessage);
  static HeartbeatResponse? _defaultInstance;
}

/// Освобождение уже отсутствующей аренды безопасно.
class ReleaseClientRequest extends $pb.GeneratedMessage {
  factory ReleaseClientRequest({
    $core.String? leaseId,
  }) {
    final result = ReleaseClientRequest._();
    if (leaseId != null) result.leaseId = leaseId;
    return result;
  }

  ReleaseClientRequest._();

  factory ReleaseClientRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReleaseClientRequest()..mergeFromBuffer(data, registry);
  factory ReleaseClientRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReleaseClientRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReleaseClientRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: ReleaseClientRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'leaseId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReleaseClientRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReleaseClientRequest copyWith(void Function(ReleaseClientRequest) updates) =>
      super.copyWith((message) => updates(message as ReleaseClientRequest))
          as ReleaseClientRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ReleaseClientRequest() / ReleaseClientRequest.new instead')
  static ReleaseClientRequest create() => ReleaseClientRequest._();
  static $pb.GeneratedMessage $_createMessage() => ReleaseClientRequest._();
  @$core.override
  ReleaseClientRequest createEmptyInstance() => ReleaseClientRequest._();
  @$core.pragma('dart2js:noInline')
  static ReleaseClientRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReleaseClientRequest>(
          ReleaseClientRequest.$_createMessage);
  static ReleaseClientRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get leaseId => $_getSZ(0);
  @$pb.TagNumber(1)
  set leaseId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasLeaseId() => $_has(0);
  @$pb.TagNumber(1)
  void clearLeaseId() => $_clearField(1);
}

class ReleaseClientResponse extends $pb.GeneratedMessage {
  factory ReleaseClientResponse() => ReleaseClientResponse._();

  ReleaseClientResponse._();

  factory ReleaseClientResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReleaseClientResponse()..mergeFromBuffer(data, registry);
  factory ReleaseClientResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ReleaseClientResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReleaseClientResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pob2.v1'),
      createEmptyInstance: ReleaseClientResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReleaseClientResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReleaseClientResponse copyWith(
          void Function(ReleaseClientResponse) updates) =>
      super.copyWith((message) => updates(message as ReleaseClientResponse))
          as ReleaseClientResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ReleaseClientResponse() / ReleaseClientResponse.new instead')
  static ReleaseClientResponse create() => ReleaseClientResponse._();
  static $pb.GeneratedMessage $_createMessage() => ReleaseClientResponse._();
  @$core.override
  ReleaseClientResponse createEmptyInstance() => ReleaseClientResponse._();
  @$core.pragma('dart2js:noInline')
  static ReleaseClientResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReleaseClientResponse>(
          ReleaseClientResponse.$_createMessage);
  static ReleaseClientResponse? _defaultInstance;
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
