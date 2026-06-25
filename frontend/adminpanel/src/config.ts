// project config

export enum MenuOrientation {
  VERTICAL = 'vertical',
  HORIZONTAL = 'horizontal'
}

export enum ThemeDirection {
  LTR = 'ltr',
  RTL = 'rtl'
}

export enum ThemeMode {
  LIGHT = 'light',
  DARK = 'dark'
}

export enum AuthProvider {
  JWT = 'jwt'
}

export const CSS_VAR_PREFIX = 'xpaywall';
export const DEFAULT_THEME_MODE = ThemeMode.DARK;
export const DASHBOARD_PATH = '/dashboard';
export const HORIZONTAL_MAX_ITEM = 6;
export const APP_AUTH: AuthProvider = AuthProvider.JWT;

const config = {
  menuOrientation: MenuOrientation.VERTICAL,
  miniDrawer: false,
  fontFamily: `'Roboto', sans-serif` as `'Roboto', sans-serif`,
  borderRadius: 8,
  outlinedFilled: true,
  presetColor: 'theme-cp0x' as const,
  i18n: 'en' as const,
  themeDirection: ThemeDirection.LTR,
  container: false
};

export default config;
