import type { FC } from "react";
import ReactJson from "react-json-view";

import { Collapse } from "antd";

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
