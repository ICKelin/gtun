export interface ApiResult<T> {
  code: number;
  message: string;
  data: T;
}

export interface LogConfig {
  Days: number;
  Level: string;
  Path: string;
}

export interface RouteConfig {
  Region: string;
  TraceAddr: string;
  Scheme: string;
  Addr: string;
  AuthKey: string;
}

export interface RegionConfig {
  Route: RouteConfig[];
  ProxyFile: string;
  Proxy: Map<string, string>;
}

export interface HttpServerConfig {
  listen_addr: string;
}

export interface Config {
  Log: LogConfig;
  Settings: Map<string, RegionConfig>;
  http_server: HttpServerConfig;
}

export interface MetaData {
  regions: string[];
  Cfg: any;
}

export interface RegionData {
  [region: string]: string[];
}

export interface MetaDataResult extends ApiResult<MetaData> {}

export default {};
