export const protocolVersion = 1 as const;

export const messageTypes = [
	'auth',
	'auth_response',
	'version_negotiate',
	'version_negotiate_ack',
	'flow_control',
	'open_tunnel',
  'open_tunnel_ack',
  'close_tunnel',
  'data',
  'heartbeat',
  'error',
  'http_request',
  'http_response',
  'tcp_data',
  'tcp_close',
  'udp_data',
  'udp_response',
] as const;

export type MessageType = (typeof messageTypes)[number];

export type ProtocolEnvelope<TPayload = unknown> = {
  version: typeof protocolVersion;
  type: MessageType;
  request_id?: string;
  payload?: TPayload;
};

export type OpenTunnel = {
  token: string;
  tunnel_id?: string;
  local_port: number;
  subdomain?: string;
  protocol: string;
  custom_domain?: string;
	password?: string;
};

export type AuthRequest = {
	token: string;
	agent_id?: string;
	requested_capabilities?: string[];
};

export type AuthResponse = {
	authenticated: boolean;
	agent_id?: string;
	organization_id?: string;
	granted_capabilities?: string[];
	error?: string;
};

export type VersionNegotiate = {
	min_version: number;
	max_version: number;
	client_name?: string;
	client_version?: string;
};

export type VersionNegotiateAck = {
	negotiated_version: number;
	supported_versions: number[];
	server_name?: string;
	server_version?: string;
};

export type FlowControl = {
	stream_id: string;
	action: 'pause' | 'resume';
	window_size?: number;
};

export type OpenTunnelAck = {
  tunnel_id: string;
  public_url: string;
  public_port?: number;
};

export type CloseTunnel = {
  tunnel_id: string;
  reason?: string;
};

export type Data = {
  tunnel_id: string;
  stream_id: string;
  data: string;
};

export type Heartbeat = {
  timestamp: number;
};

export type ErrorMessage = {
  code: string;
  message: string;
};

export type HTTPRequest = {
  method: string;
  path: string;
  headers: Record<string, string[]>;
  body?: string;
};

export type HTTPResponse = {
  status_code: number;
  headers: Record<string, string[]>;
  body?: string;
  error?: string;
};

export type TCPData = {
  tunnel_id?: string;
  connection_id: string;
  data: string;
};

export type TCPClose = {
  tunnel_id?: string;
  connection_id: string;
  reason?: string;
};

export type UDPData = {
  tunnel_id?: string;
  packet_id: string;
  source_address: string;
  source_port: number;
  data: string;
};

export type UDPResponse = {
  tunnel_id?: string;
  packet_id: string;
  target_address: string;
  target_port: number;
  data: string;
};

export const encodeMessage = <TPayload>(
  message: ProtocolEnvelope<TPayload>,
): string => JSON.stringify(message);

export const decodeMessage = <TPayload>(
  value: string,
): ProtocolEnvelope<TPayload> => {
  const message = JSON.parse(value) as ProtocolEnvelope<TPayload>;
  if (message.version !== protocolVersion || !messageTypes.includes(message.type)) {
    throw new Error('unsupported tunnel protocol message');
  }
  return message;
};
