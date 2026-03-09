import type { TableProps as TablePropsA } from "antd";
import { Table as TableA } from "antd";

export interface TableProps<RecordType> extends TablePropsA<RecordType> {}
export const Table = <RecordType extends object = any>({ ...props }: TableProps<RecordType>) => {
  return <TableA bordered size="small" {...props} />;
};
