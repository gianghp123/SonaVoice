import { IWebRTCConnection } from "@/lib/types/webtc-connection.interface";

export interface ICreateSessionRes {
  id: string;
  webrtcConnection: IWebRTCConnection;
}