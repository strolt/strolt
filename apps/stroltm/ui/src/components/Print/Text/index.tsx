import { FC } from "react";

import { Typography } from "antd";

export interface TextProps {
  value?: string;
  copyable?: boolean;
}
export const Text: FC<TextProps> = ({ value, copyable }) => {
  if (!value) {
    return <>-</>;
  }

  return <Typography.Text copyable={copyable}>{value}</Typography.Text>;
};
