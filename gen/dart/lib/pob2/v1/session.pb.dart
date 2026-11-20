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
  }) {
    final result = GetServerInfoResponse._();
    if (version != null) result.version = version;
    if (protocolVersion != null) result.protocolVersion = protocolVersion;
    if (capabilities != null) result.capabilities.addAll(capabilities);
    if (state != null) result.state = state;
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

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
