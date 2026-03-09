import { Divider, Space } from "antd";

import type { TaskListItemDestination } from "stores/manager.store/taskList";

import { Link } from "components";

interface LinksProps {
  destinationName?: string;
  instanceName?: string;
  proxyName?: string;
  serviceName?: string;
  taskName?: string;
}
const Links: React.FC<LinksProps> = (params) => {
  if (!params.instanceName && !params.serviceName && !params.taskName && params.destinationName) {
    return null;
  }

  const linkParams = {
    destinationId: params.destinationName,
    instanceId: params.instanceName,
    proxyId: params.proxyName,
    serviceId: params.serviceName,
    taskId: params.taskName,
  };

  return (
    <Space wrap>
      <Link
        params={linkParams}
        style={{ display: "block" }}
        to={
          !!linkParams.proxyId
            ? "instances.proxyId.instanceId.serviceId.taskId.destinationId.proxySnapshotList"
            : "instances.instanceId.serviceId.taskId.destinationId.snapshotList"
        }
      >
        Snapshots
      </Link>

      <Divider style={{ margin: 0 }} type="vertical" />

      <Link
        params={linkParams}
        style={{ display: "block" }}
        to={
          !!linkParams.proxyId
            ? "instances.proxyId.instanceId.serviceId.taskId.destinationId.prune"
            : "instances.instanceId.serviceId.taskId.destinationId.prune"
        }
      >
        Prune
      </Link>

      <Divider style={{ margin: 0 }} type="vertical" />

      <Link
        params={linkParams}
        style={{ display: "block" }}
        to={
          !!linkParams.proxyId
            ? "instances.proxyId.instanceId.serviceId.taskId.destinationId.proxyStats"
            : "instances.instanceId.serviceId.taskId.destinationId.stats"
        }
      >
        Stats
      </Link>
    </Space>
  );
};

export interface TaskListItemDestinationProps extends LinksProps, TaskListItemDestination {}
export const PrintDestination: React.FC<TaskListItemDestinationProps> = (el) => {
  return (
    <Space direction="vertical">
      <Space wrap>
        <span>{el.name}:</span>
        <b>{el.driver}</b>
      </Space>

      <Links
        destinationName={el.destinationName}
        instanceName={el.instanceName}
        proxyName={el.proxyName}
        serviceName={el.serviceName}
        taskName={el.taskName}
      />
    </Space>
  );
};

export interface PrintDestinationsProps {
  list: TaskListItemDestinationProps[];
}
export const PrintDestinations: React.FC<PrintDestinationsProps> = ({ list }) => {
  if (!list.length) {
    return <>-</>;
  }

  return (
    <>
      {list.map((el) => (
        <Space direction="vertical" key={`${el.name}_${el.driver}`}>
          <PrintDestination {...el} />
        </Space>
      ))}
    </>
  );
};
