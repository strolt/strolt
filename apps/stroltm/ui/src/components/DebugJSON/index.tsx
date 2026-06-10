import type { FC } from "react";

import { Collapse } from "antd";

import ReactJson from "@microlink/react-json-view";

export interface DebugJSONProps {
  data: any;
  title?: string;
}
export const DebugJSON: FC<DebugJSONProps> = ({ data, title }) => {
  return (
    <Collapse>
      <Collapse.Panel header={title || "raw"} key="raw">
        <ReactJson src={data || {}} />
      </Collapse.Panel>
    </Collapse>
  );
};
