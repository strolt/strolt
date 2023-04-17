import { Divider, Space } from "antd";

import { Link } from "components";

import { TaskListItemDestination } from "stores/manager.store/taskList";

interface LinksProps {
  proxyName?: string;
  instanceName?: string;
  serviceName?: string;
  taskName?: string;
  destinationName?: string;
}
const Links: React.FC<LinksProps> = (params) => {
  if (!params.instanceName && !params.serviceName && !params.taskName && params.destinationName) {
    return null;
  }

  const linkParams = {
    proxyId: params.proxyName,
    instanceId: params.instanceName,
    serviceId: params.serviceName,
    taskId: params.taskName,
    destinationId: params.destinationName,
  };

  return (
    <Space wrap>
      <Link
        to={
          !!linkParams.proxyId
            ? "instances.proxyId.instanceId.serviceId.taskId.destinationId.proxySnapshotList"
            : "instances.instanceId.serviceId.taskId.destinationId.snapshotList"
        }
        params={linkParams}
        style={{ display: "block" }}
      >
        Snapshots
      </Link>

      <Divider type="vertical" style={{ margin: 0 }} />

      <Link
        to={
          !!linkParams.proxyId
            ? "instances.proxyId.instanceId.serviceId.taskId.destinationId.prune"
            : "instances.instanceId.serviceId.taskId.destinationId.prune"
        }
        params={linkParams}
        style={{ display: "block" }}
      >
        Prune
      </Link>

      <Divider type="vertical" style={{ margin: 0 }} />

      <Link
        to={
          !!linkParams.proxyId
            ? "instances.proxyId.instanceId.serviceId.taskId.destinationId.proxyStats"
            : "instances.instanceId.serviceId.taskId.destinationId.stats"
        }
        params={linkParams}
        style={{ display: "block" }}
      >
        Stats
      </Link>
    </Space>
  );
};

export interface TaskListItemDestinationProps extends TaskListItemDestination, LinksProps {}
export const PrintDestination: React.FC<TaskListItemDestinationProps> = (el) => {
  return (
    <Space direction="vertical">
      <Space wrap>
        <span>{el.name}:</span>
        <b>{el.driver}</b>
      </Space>

      <Links
        proxyName={el.proxyName}
        instanceName={el.instanceName}
        serviceName={el.serviceName}
        taskName={el.taskName}
        destinationName={el.destinationName}
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
