export type ServiceInstanceState = 'running' | 'stopped' | 'none';

export interface PiniaConfigStore {
  app?: ConfigVersionApp;
}

export interface ConfigVersionApp {
  app: ConfigArchInfoApp[];
}

export type ConfigOSApp = 'Windows' | 'Linux' | 'macOS';

export interface ConfigArchInfoApp {
  os: ConfigOSApp;
  data: ConfigDataApp[];
}

export interface ConfigDataApp {
  version: string;
  gpg?: string;
  link: string;
}
