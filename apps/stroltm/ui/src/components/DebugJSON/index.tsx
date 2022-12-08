import { FC } from "react";

import ReactJson from "react-json-view";

import { Collapse } from "antd";

export interface DebugJSONProps {
  data: any;
  title?: string;
}
export const DebugJSON: FC<DebugJSONProps> = ({ data, title }) => {
  return (
    <Collapse>
      <Collapse.Panel key="raw" header={title || "raw"}>
        <ReactJson src={data || {}} />
      </Collapse.Panel>
    </Collapse>
  );
};
