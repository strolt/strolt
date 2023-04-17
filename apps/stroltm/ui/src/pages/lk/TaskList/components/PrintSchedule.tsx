import cronstrue from "cronstrue";

import { Popover, Space } from "antd";

import { TaskListItemSchedule } from "stores/manager.store/taskList";

interface PrintCronProps {
  schedule: string;
}
const PrintCron: React.FC<PrintCronProps> = ({ schedule }) => {
  return (
    <Popover content={() => cronstrue.toString(schedule)}>
      <b>"{schedule}"</b>
    </Popover>
  );
};

export const PrintSchedule: React.FC<TaskListItemSchedule> = (el) => {
  if (!el.backup && !el.prune) {
    return <>-</>;
  }

  return (
    <>
      {!!el.backup && (
        <div style={{ display: "flex", justifyContent: "space-between" }}>
          <span>backup:</span>
          <PrintCron schedule={el.backup} />
        </div>
      )}
      {!!el.prune && (
        <div style={{ display: "flex", justifyContent: "space-between" }}>
          <span>prune:</span>
          <PrintCron schedule={el.prune} />
        </div>
      )}
    </>
  );
};
