import { Divider, Space } from "antd";

import { Print } from "components";

import { TaskListItemNotification } from "stores/manager.store/taskList";

export const PrintNotification: React.FC<TaskListItemNotification> = (el) => {
  return (
    <Space direction="vertical">
      <Space wrap>
        <span>{el.name}:</span>
        <b>{el.driver}</b>
      </Space>
      <Print.TagList value={el.events} />
    </Space>
  );
};

export interface PrintNotificationsProps {
  list: TaskListItemNotification[];
}
export const PrintNotifications: React.FC<PrintNotificationsProps> = ({ list }) => {
  if (!list.length) {
    return <>-</>;
  }

  return (
    <>
      {list.map((el, i) => (
        <div key={`${el.driver}_${el.name}`}>
          <PrintNotification {...el} />
          {list.length - 1 !== i && <Divider style={{ margin: "0.5rem 0" }} />}
        </div>
      ))}
    </>
  );
};
