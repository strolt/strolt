import {
  CheckSquareFilled,
  CheckCircleFilled,
  CloseSquareFilled,
  CloseCircleFilled,
} from "@ant-design/icons";

import { FC } from "react";

const iconCheck = (square?: boolean, size?: number | string) => {
  if (square) {
    return (
      <CheckSquareFilled
        style={{
          fontSize: size,
          color: "green",
        }}
      />
    );
  }

  return (
    <CheckCircleFilled
      style={{
        fontSize: size,
        color: "green",
      }}
    />
  );
};

const iconUnCheck = (square?: boolean, size?: number | string) => {
  if (square) {
    return (
      <CloseSquareFilled
        style={{
          fontSize: size,
          color: "red",
        }}
      />
    );
  }

  return (
    <CloseCircleFilled
      style={{
        fontSize: size,
        color: "red",
      }}
    />
  );
};

export interface BooleanProps {
  value?: string | boolean | number;
  square?: boolean;
  size?: number | string;
}
export const Boolean: FC<BooleanProps> = ({ value, square, size }) => {
  if (!value) {
    return iconUnCheck(square, size);
  }

  return iconCheck(square, size);
};
