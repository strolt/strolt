import type { FC } from "react";

import {
  CheckCircleFilled,
  CheckSquareFilled,
  CloseCircleFilled,
  CloseSquareFilled,
} from "@ant-design/icons";

const iconCheck = (square?: boolean, size?: number | string) => {
  if (square) {
    return (
      <CheckSquareFilled
        style={{
          color: "green",
          fontSize: size,
        }}
      />
    );
  }

  return (
    <CheckCircleFilled
      style={{
        color: "green",
        fontSize: size,
      }}
    />
  );
};

const iconUnCheck = (square?: boolean, size?: number | string) => {
  if (square) {
    return (
      <CloseSquareFilled
        style={{
          color: "red",
          fontSize: size,
        }}
      />
    );
  }

  return (
    <CloseCircleFilled
      style={{
        color: "red",
        fontSize: size,
      }}
    />
  );
};

export interface BooleanProps {
  size?: number | string;
  square?: boolean;
  value?: boolean | number | string;
}
export const Boolean: FC<BooleanProps> = ({ size, square, value }) => {
  if (!value) {
    return iconUnCheck(square, size);
  }

  return iconCheck(square, size);
};
