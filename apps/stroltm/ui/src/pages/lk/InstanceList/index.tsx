import type { FC } from "react";
import { useEffect } from "react";

import { Button, message, Popconfirm, Tag, Typography } from "antd";

import InternalCard from "antd/es/card/Card";

const Card = InternalCard;

import type { ConfigServiceTask, ManagerPreparedInstance } from "api/generated";

import { DebugJSON, LatestVersionLink, Link, TagColored } from "components";
import { observer, useStores } from "stores";

export interface BackupButtonProps {
  instanceName: string;
  proxyName?: string;
  serviceName: string;
  taskName: string;
}
const BackupButton: FC<BackupButtonProps> = observer(
  ({ instanceName, proxyName, serviceName, taskName }) => {
    const { managerStore } = useStores();

    const status = managerStore.backupStatusMap.get(
      managerStore.backupStatusMapKey(instanceName, serviceName, taskName, proxyName),
    );

    return (
      <Popconfirm
        okText="Yes"
        onConfirm={() => managerStore.backup(instanceName, serviceName, taskName, proxyName)}
        title="Are you sure?"
      >
        <Button
          danger
          loading={
            status?.state === "pending" ||
            managerStore.taskStatusMap.get(
              managerStore.getTaskStatusMapKey(instanceName, serviceName, taskName, proxyName),
            )?.isRunning
          }
          size="small"
        >
          Backup
        </Button>
      </Popconfirm>
    );
  },
);

const BackupAll: FC = observer(() => {
  const { managerStore } = useStores();

  useEffect(() => {
    return () => managerStore.resetBackupAll();
  }, []);

  const handleClick = async () => {
    const data = await managerStore.backupAll();
    message.info(`Success started: ${data.successStarted?.length}`);
    message.error(
      `Error started: ${data.errorStarted?.length} [${data.errorStarted
        ?.map((el) => [el.instanceName, el.serviceName, el.taskName].filter(Boolean).join(" "))
        .filter(Boolean)
        .join(", ")}]`,
    );
  };

  return (
    <Popconfirm okText="Yes" onConfirm={handleClick} title="Are you sure?">
      <Button
        danger
        loading={managerStore.backupAllStatus?.state === "pending"}
        style={{ marginBottom: "1rem" }}
        type="primary"
      >
        Backup ALL
      </Button>
    </Popconfirm>
  );
});

export interface TaskProps {
  instanceName: string;
  proxyName?: string;
  serviceName: string;
  task: ConfigServiceTask;
  taskName: string;
}
const Task: FC<TaskProps> = observer(({ instanceName, proxyName, serviceName, task, taskName }) => {
  return (
    <Card
      extra={
        <BackupButton
          instanceName={instanceName}
          proxyName={proxyName}
          serviceName={serviceName}
          taskName={taskName}
        />
      }
      size="small"
      style={{ marginBottom: "1rem" }}
      title={`task: [${taskName}]`}
    >
      <div>
        TAGS:
        {task?.tags?.length ? (
          task?.tags.map((tag) => <TagColored key={tag} value={tag} />)
        ) : (
          <b>-</b>
        )}
      </div>

      <div>
        Schedule:
        <ul>
          <li>
            backup: <b>{task.schedule?.backup || "-"}</b>
          </li>
          <li>
            prune: <b>{task.schedule?.prune || "-"}</b>
          </li>
        </ul>
      </div>

      <div>
        Destinations ({Object.entries(task.destinations || {}).length}):
        <ul>
          {Object.entries(task.destinations || {}).map(([destinationName, destination]) => {
            return (
              <li key={destinationName}>
                {destinationName}: <b>{destination.driver}</b>
                {" | "}
                <Link
                  params={{
                    destinationId: destinationName,
                    instanceId: instanceName,
                    proxyId: proxyName,
                    serviceId: serviceName,
                    taskId: taskName,
                  }}
                  to={
                    proxyName
                      ? "instances.proxyId.instanceId.serviceId.taskId.destinationId.proxySnapshotList"
                      : "instances.instanceId.serviceId.taskId.destinationId.snapshotList"
                  }
                >
                  Snapshots
                </Link>
                {" | "}
                <Link
                  params={{
                    destinationId: destinationName,
                    instanceId: instanceName,
                    serviceId: serviceName,
                    taskId: taskName,
                  }}
                  to="instances.instanceId.serviceId.taskId.destinationId.prune"
                >
                  Prune
                </Link>
                {" | "}
                <Link
                  params={{
                    destinationId: destinationName,
                    instanceId: instanceName,
                    serviceId: serviceName,
                    taskId: taskName,
                  }}
                  to="instances.instanceId.serviceId.taskId.destinationId.stats"
                >
                  Stats
                </Link>
              </li>
            );
          })}
        </ul>
      </div>

      <div>
        Notifications ({task.notifications?.length || 0}):
        {task.notifications?.length ? (
          <ul>
            {task.notifications.map((notification) => (
              <li key={notification.name}>
                {notification.name} (<b>{notification.driver}</b>):{" "}
                {notification.events?.map((eventName) => (
                  <TagColored key={eventName} value={eventName} />
                ))}
              </li>
            ))}
          </ul>
        ) : (
          <b>-</b>
        )}
      </div>
    </Card>
  );
});

export interface ServiceProps {
  instanceName: string;
  proxyName?: string;
  service: Record<string, ConfigServiceTask>;
  serviceName: string;
}
const Service: FC<ServiceProps> = observer(({ instanceName, proxyName, service, serviceName }) => {
  return (
    <Card size="small" style={{ marginBottom: "1rem" }} title={`service: [${serviceName}]`}>
      {Object.entries(service).map(([taskName, task]) => {
        return (
          <Task
            instanceName={instanceName}
            key={taskName}
            proxyName={proxyName}
            serviceName={serviceName}
            task={task}
            taskName={taskName}
          />
        );
      })}
    </Card>
  );
});

export interface InstanceProps {
  instance: ManagerPreparedInstance;
}
const Instance: FC<InstanceProps> = observer(({ instance }) => {
  const { infoStore } = useStores();

  const instanceInfo = infoStore.map.get(instance.name || "");

  return (
    <div style={{ minWidth: "25rem" }}>
      <Card
        extra={
          instanceInfo?.isOnline ? (
            <Tag color="success">Online</Tag>
          ) : (
            <Tag color="error">Offline</Tag>
          )
        }
        size="small"
        title={
          <>
            {[
              Boolean(instance.proxyName) && `proxy: [${instance.proxyName}]`,
              `instance: [${instance.name}]`,
              `version: ${instanceInfo?.version}`,
            ]
              .filter(Boolean)
              .join(" ")}{" "}
            <LatestVersionLink version={instanceInfo?.version} /> ({instance.config?.timezone})
          </>
        }
      >
        <DebugJSON data={instanceInfo || {}} />

        {Object.entries(instance.config?.services || {}).map(([serviceName, service]) => {
          return (
            <>
              <div>
                TAGS:
                {instance.config?.tags?.length ? (
                  instance.config?.tags.map((tag) => <TagColored key={tag} value={tag} />)
                ) : (
                  <b>-</b>
                )}
              </div>

              <Service
                instanceName={instance.name || ""}
                key={serviceName}
                proxyName={instance.proxyName}
                service={service}
                serviceName={serviceName}
              />
            </>
          );
        })}
      </Card>
    </div>
  );
});

const InstanceList = observer(() => {
  const { managerStore } = useStores();

  return (
    <div>
      <Typography.Title>Instances:</Typography.Title>
      <BackupAll />
      <div style={{ display: "flex", flexWrap: "wrap", gap: "1rem" }}>
        {managerStore.instances.map((instance) => {
          return <Instance instance={instance} key={`${instance.proxyName}_${instance.name}`} />;
        })}
      </div>

      <DebugJSON data={managerStore.instances} title="Instances" />
    </div>
  );
});

export default InstanceList;
