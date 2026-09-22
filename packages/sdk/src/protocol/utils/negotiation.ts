import type {
  VersionNegotiate,
  VersionNegotiateAck,
} from '../interfaces/messages';
import {
  maxSupportedVersion,
  minSupportedVersion,
  protocolVersion,
} from './constants';

export const negotiateVersion = (
  request: VersionNegotiate,
): VersionNegotiateAck => {
  if (
    request.max_version < minSupportedVersion ||
    request.min_version > maxSupportedVersion
  ) {
    throw new Error(
      `incompatible protocol version request: ${request.min_version}-${request.max_version}`,
    );
  }
  return {
    negotiated_version: Math.min(request.max_version, maxSupportedVersion),
    supported_versions: [protocolVersion],
    server_name: 'outpipe',
    server_version: '0.1.0',
  };
};
