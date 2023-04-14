import { observer } from "mobx-react-lite";

import { ConfigProvider, theme } from "antd";
import { AliasToken } from "antd/es/theme/internal";

import { appConfigStore } from "stores/app-config.store";

const themeTokenCommon: Partial<AliasToken> = {
  colorPrimary: "#00b96b",
};

const themeTokenDark: Partial<AliasToken> = {
  ...themeTokenCommon,
  colorBgBase: "#1c2128",
};

const themeTokenLight: Partial<AliasToken> = {
  ...themeTokenCommon,
};

export const ThemeProvider = observer(({ children }: { children: React.ReactNode }) => {
  return (
    <ConfigProvider
      theme={{
        algorithm: appConfigStore.mode === "dark" ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: appConfigStore.mode === "dark" ? themeTokenDark : themeTokenLight,
      }}
    >
      {children}
    </ConfigProvider>
  );
});
