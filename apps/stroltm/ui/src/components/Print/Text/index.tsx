import type { FC } from "react";

import { Typography } from "antd";

export interface TextProps {
  copyable?: boolean;
  value?: string;
}
export const Text: FC<TextProps> = ({ copyable, value }) => {
  if (!value) {
    return <>-</>;
  }

  return <Typography.Text copyable={copyable}>{value}</Typography.Text>;
};
