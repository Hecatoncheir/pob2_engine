// This is a generated file - do not edit.
//
// Generated from pob2/v1/session.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'package:protobuf/protobuf.dart' as $pb;

import 'session.pb.dart' as $0;

export 'session.pb.dart';

/// SessionService предоставляет общий сервис сессий первого этапа.
@$pb.GrpcServiceName('pob2.v1.SessionService')
class SessionServiceClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  SessionServiceClient(super.channel, {super.options, super.interceptors});

  $grpc.ResponseFuture<$0.GetServerInfoResponse> getServerInfo(
    $0.GetServerInfoRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getServerInfo, request, options: options);
  }

  $grpc.ResponseFuture<$0.CreateSessionResponse> createSession(
    $0.CreateSessionRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createSession, request, options: options);
  }

  $grpc.ResponseFuture<$0.GetSnapshotResponse> getSnapshot(
    $0.GetSnapshotRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getSnapshot, request, options: options);
  }

  $grpc.ResponseFuture<$0.RenameBuildResponse> renameBuild(
    $0.RenameBuildRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$renameBuild, request, options: options);
  }

  $grpc.ResponseFuture<$0.ShutdownResponse> shutdown(
    $0.ShutdownRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$shutdown, request, options: options);
  }

  $grpc.ResponseFuture<$0.RegisterClientResponse> registerClient(
    $0.RegisterClientRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$registerClient, request, options: options);
  }

  $grpc.ResponseFuture<$0.HeartbeatResponse> heartbeat(
    $0.HeartbeatRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$heartbeat, request, options: options);
  }

  $grpc.ResponseFuture<$0.ReleaseClientResponse> releaseClient(
    $0.ReleaseClientRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$releaseClient, request, options: options);
  }

  // method descriptors

  static final _$getServerInfo =
      $grpc.ClientMethod<$0.GetServerInfoRequest, $0.GetServerInfoResponse>(
          '/pob2.v1.SessionService/GetServerInfo',
          ($0.GetServerInfoRequest value) => value.writeToBuffer(),
          $0.GetServerInfoResponse.fromBuffer);
  static final _$createSession =
      $grpc.ClientMethod<$0.CreateSessionRequest, $0.CreateSessionResponse>(
          '/pob2.v1.SessionService/CreateSession',
          ($0.CreateSessionRequest value) => value.writeToBuffer(),
          $0.CreateSessionResponse.fromBuffer);
  static final _$getSnapshot =
      $grpc.ClientMethod<$0.GetSnapshotRequest, $0.GetSnapshotResponse>(
          '/pob2.v1.SessionService/GetSnapshot',
          ($0.GetSnapshotRequest value) => value.writeToBuffer(),
          $0.GetSnapshotResponse.fromBuffer);
  static final _$renameBuild =
      $grpc.ClientMethod<$0.RenameBuildRequest, $0.RenameBuildResponse>(
          '/pob2.v1.SessionService/RenameBuild',
          ($0.RenameBuildRequest value) => value.writeToBuffer(),
          $0.RenameBuildResponse.fromBuffer);
  static final _$shutdown =
      $grpc.ClientMethod<$0.ShutdownRequest, $0.ShutdownResponse>(
          '/pob2.v1.SessionService/Shutdown',
          ($0.ShutdownRequest value) => value.writeToBuffer(),
          $0.ShutdownResponse.fromBuffer);
  static final _$registerClient =
      $grpc.ClientMethod<$0.RegisterClientRequest, $0.RegisterClientResponse>(
          '/pob2.v1.SessionService/RegisterClient',
          ($0.RegisterClientRequest value) => value.writeToBuffer(),
          $0.RegisterClientResponse.fromBuffer);
  static final _$heartbeat =
      $grpc.ClientMethod<$0.HeartbeatRequest, $0.HeartbeatResponse>(
          '/pob2.v1.SessionService/Heartbeat',
          ($0.HeartbeatRequest value) => value.writeToBuffer(),
          $0.HeartbeatResponse.fromBuffer);
  static final _$releaseClient =
      $grpc.ClientMethod<$0.ReleaseClientRequest, $0.ReleaseClientResponse>(
          '/pob2.v1.SessionService/ReleaseClient',
          ($0.ReleaseClientRequest value) => value.writeToBuffer(),
          $0.ReleaseClientResponse.fromBuffer);
}

@$pb.GrpcServiceName('pob2.v1.SessionService')
abstract class SessionServiceBase extends $grpc.Service {
  $core.String get $name => 'pob2.v1.SessionService';

  SessionServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$0.GetServerInfoRequest, $0.GetServerInfoResponse>(
            'GetServerInfo',
            getServerInfo_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.GetServerInfoRequest.fromBuffer(value),
            ($0.GetServerInfoResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.CreateSessionRequest, $0.CreateSessionResponse>(
            'CreateSession',
            createSession_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.CreateSessionRequest.fromBuffer(value),
            ($0.CreateSessionResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.GetSnapshotRequest, $0.GetSnapshotResponse>(
            'GetSnapshot',
            getSnapshot_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.GetSnapshotRequest.fromBuffer(value),
            ($0.GetSnapshotResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.RenameBuildRequest, $0.RenameBuildResponse>(
            'RenameBuild',
            renameBuild_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.RenameBuildRequest.fromBuffer(value),
            ($0.RenameBuildResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ShutdownRequest, $0.ShutdownResponse>(
        'Shutdown',
        shutdown_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.ShutdownRequest.fromBuffer(value),
        ($0.ShutdownResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.RegisterClientRequest,
            $0.RegisterClientResponse>(
        'RegisterClient',
        registerClient_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.RegisterClientRequest.fromBuffer(value),
        ($0.RegisterClientResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.HeartbeatRequest, $0.HeartbeatResponse>(
        'Heartbeat',
        heartbeat_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.HeartbeatRequest.fromBuffer(value),
        ($0.HeartbeatResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.ReleaseClientRequest, $0.ReleaseClientResponse>(
            'ReleaseClient',
            releaseClient_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.ReleaseClientRequest.fromBuffer(value),
            ($0.ReleaseClientResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.GetServerInfoResponse> getServerInfo_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.GetServerInfoRequest> $request) async {
    return getServerInfo($call, await $request);
  }

  $async.Future<$0.GetServerInfoResponse> getServerInfo(
      $grpc.ServiceCall call, $0.GetServerInfoRequest request);

  $async.Future<$0.CreateSessionResponse> createSession_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.CreateSessionRequest> $request) async {
    return createSession($call, await $request);
  }

  $async.Future<$0.CreateSessionResponse> createSession(
      $grpc.ServiceCall call, $0.CreateSessionRequest request);

  $async.Future<$0.GetSnapshotResponse> getSnapshot_Pre($grpc.ServiceCall $call,
      $async.Future<$0.GetSnapshotRequest> $request) async {
    return getSnapshot($call, await $request);
  }

  $async.Future<$0.GetSnapshotResponse> getSnapshot(
      $grpc.ServiceCall call, $0.GetSnapshotRequest request);

  $async.Future<$0.RenameBuildResponse> renameBuild_Pre($grpc.ServiceCall $call,
      $async.Future<$0.RenameBuildRequest> $request) async {
    return renameBuild($call, await $request);
  }

  $async.Future<$0.RenameBuildResponse> renameBuild(
      $grpc.ServiceCall call, $0.RenameBuildRequest request);

  $async.Future<$0.ShutdownResponse> shutdown_Pre($grpc.ServiceCall $call,
      $async.Future<$0.ShutdownRequest> $request) async {
    return shutdown($call, await $request);
  }

  $async.Future<$0.ShutdownResponse> shutdown(
      $grpc.ServiceCall call, $0.ShutdownRequest request);

  $async.Future<$0.RegisterClientResponse> registerClient_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.RegisterClientRequest> $request) async {
    return registerClient($call, await $request);
  }

  $async.Future<$0.RegisterClientResponse> registerClient(
      $grpc.ServiceCall call, $0.RegisterClientRequest request);

  $async.Future<$0.HeartbeatResponse> heartbeat_Pre($grpc.ServiceCall $call,
      $async.Future<$0.HeartbeatRequest> $request) async {
    return heartbeat($call, await $request);
  }

  $async.Future<$0.HeartbeatResponse> heartbeat(
      $grpc.ServiceCall call, $0.HeartbeatRequest request);

  $async.Future<$0.ReleaseClientResponse> releaseClient_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.ReleaseClientRequest> $request) async {
    return releaseClient($call, await $request);
  }

  $async.Future<$0.ReleaseClientResponse> releaseClient(
      $grpc.ServiceCall call, $0.ReleaseClientRequest request);
}
