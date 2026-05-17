export interface IWebRTCConnection {
  sessionId: string;
  iceConfig?: IceConfig;
}

export interface IceConfig {
  iceServers: IceServer[];
}

export interface IceServer {
  urls: string[];
  username?: string;
  credential?: string;
}