import { FC, ReactNode } from "react";

import { Spin } from "antd";

export interface LoadingHOCProps {
  children: ReactNode;
  loading: boolean;
}

export const LoadingHOC: FC<LoadingHOCProps> = ({ children, loading }) => {
  if (loading) {
    return <Spin />;
  }

  return <>{children}</>;
};
