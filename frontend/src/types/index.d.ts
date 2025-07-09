export type ServiceInstanceState = 'running' | 'stopped' | 'none';

export interface PiniaConfigStore {
  mysql?: ConfigVersionMySQL;
}

export interface ConfigVersionMySQL {
  mysql: ConfigArchInfoMySQL[];
}

export type ConfigOSMySQL = 'Windows' | 'Linux' | 'macOS';

export interface ConfigArchInfoMySQL {
  os: ConfigOSMySQL;
  data: ConfigDataMySQL[];
}

export interface ConfigDataMySQL {
  version: string;
  gpg?: string;
  link: string;
}
