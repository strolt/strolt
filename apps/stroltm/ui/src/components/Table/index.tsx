import { TableProps as TablePropsA, Table as TableA } from "antd";

export interface TableProps<RecordType> extends TablePropsA<RecordType> {}
export const Table = <RecordType extends object = any>({ ...props }: TableProps<RecordType>) => {
  return <TableA size="small" bordered {...props} />;
};
