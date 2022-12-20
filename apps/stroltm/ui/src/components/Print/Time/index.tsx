import { FC } from "react";

export interface TimeProps {
  value?: string | Date;
  withTime?: boolean;
}
export const Time: FC<TimeProps> = ({ value, withTime }) => {
  if (!value) {
    return <>-</>;
  }

  const date = new Date(value);

  if (withTime) {
    return (
      <>
        {new Intl.DateTimeFormat("default", { dateStyle: "short", timeStyle: "medium" }).format(
          date,
        )}
      </>
    );
  }

  return <>{new Intl.DateTimeFormat("default", { dateStyle: "short" }).format(date)}</>;
};
