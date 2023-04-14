import { FC, useEffect } from "react";

import { Button, Card, message, Popconfirm, Tag, Typography } from "antd";

import { DebugJSON, LatestVersionLink, Link, TagColored } from "components";

import {
  ManagerhManagerPreparedInstance,
  StroltpModelsStroltModelsAPIConfigServiceTask,
} from "api/generated";

import { observer, useStores } from "stores";

export interface BackupButtonProps {
  proxyName?: string;
  instanceName: string;
  serviceName: string;
  taskName: string;
}
const BackupButton: FC<BackupButtonProps> = observer(
  ({ proxyName, instanceName, serviceName, taskName }) => {
    const { managerStore } = useStores();

    const status = managerStore.backupStatusMap.get(
      managerStore.backupStatusMapKey(instanceName, serviceName, taskName, proxyName),
    );

    return (
      <Popconfirm
        title="Are you sure?"
        onConfirm={() => managerStore.backup(instanceName, serviceName, taskName, proxyName)}
      >
        <Button
          loading={
            status?.state === "pending" ||
            managerStore.taskStatusMap.get(
              managerStore.getTaskStatusMapKey(instanceName, serviceName, taskName, proxyName),
            )?.isRunning
          }
          size="small"
          danger
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
    <Popconfirm title="Are you sure?" onConfirm={handleClick}>
      <Button
        type="primary"
        style={{ marginBottom: "1rem" }}
        loading={managerStore.backupAllStatus?.state === "pending"}
        danger
      >
        Backup ALL
      </Button>
    </Popconfirm>
  );
});

export interface TaskProps {
  proxyName?: string;
  instanceName: string;
  serviceName: string;
  taskName: string;
  task: StroltpModelsStroltModelsAPIConfigServiceTask;
}
const Task: FC<TaskProps> = observer(({ proxyName, task, instanceName, serviceName, taskName }) => {
  return (
    <Card
      size="small"
      style={{ marginBottom: "1rem" }}
      title={`task: [${taskName}]`}
      extra={
        <BackupButton
          proxyName={proxyName}
          instanceName={instanceName}
          serviceName={serviceName}
          taskName={taskName}
        />
      }
    >
      <div>
        TAGS:
        {task?.tags?.length ? task?.tags.map((tag) => <TagColored key={tag} value={tag}/>) : <b>-</b>}
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
                  to={
                    !!proxyName
                      ? "instances.proxyId.instanceId.serviceId.taskId.destinationId.proxySnapshotList"
                      : "instances.instanceId.serviceId.taskId.destinationId.snapshotList"
                  }
                  params={{
                    proxyId: proxyName,
                    instanceId: instanceName,
                    serviceId: serviceName,
                    taskId: taskName,
                    destinationId: destinationName,
                  }}
                >
                  Snapshots
                </Link>
                {" | "}
                <Link
                  to="instances.instanceId.serviceId.taskId.destinationId.prune"
                  params={{
                    instanceId: instanceName,
                    serviceId: serviceName,
                    taskId: taskName,
                    destinationId: destinationName,
                  }}
                >
                  Prune
                </Link>
                {" | "}
                <Link
                  to="instances.instanceId.serviceId.taskId.destinationId.stats"
                  params={{
                    instanceId: instanceName,
                    serviceId: serviceName,
                    taskId: taskName,
                    destinationId: destinationName,
                  }}
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
                  <TagColored key={eventName} value={eventName}/>
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
  proxyName?: string;
  instanceName: string;
  serviceName: string;
  service: Record<string, StroltpModelsStroltModelsAPIConfigServiceTask>;
}
const Service: FC<ServiceProps> = observer(({ proxyName, service, serviceName, instanceName }) => {
  return (
    <Card size="small" style={{ marginBottom: "1rem" }} title={`service: [${serviceName}]`}>
      {Object.entries(service).map(([taskName, task]) => {
        return (
          <Task
            key={taskName}
            proxyName={proxyName}
            instanceName={instanceName}
            serviceName={serviceName}
            taskName={taskName}
            task={task}
          />
        );
      })}
    </Card>
  );
});

export interface InstanceProps {
  instance: ManagerhManagerPreparedInstance;
}
const Instance: FC<InstanceProps> = observer(({ instance }) => {
  const { infoStore } = useStores();

  const instanceInfo = infoStore.map.get(instance.name || "");

  return (
    <div style={{ minWidth: "25rem" }}>
      <Card
        size="small"
        title={
          <>
            {[
              !!instance.proxyName && `proxy: [${instance.proxyName}]`,
              `instance: [${instance.name}]`,
              `version: ${instanceInfo?.version}`,
            ]
              .filter(Boolean)
              .join(" ")}{" "}
            <LatestVersionLink version={instanceInfo?.version} /> ({instance.config?.timezone})
          </>
        }
        extra={
          instanceInfo?.isOnline ? (
            <Tag color="success">Online</Tag>
          ) : (
            <Tag color="error">Offline</Tag>
          )
        }
      >
        <DebugJSON data={instanceInfo || {}} />

        {Object.entries(instance.config?.services || {}).map(([serviceName, service]) => {
          return (
            <>
              <div>
                TAGS:
                {instance.config?.tags?.length ? (
                  instance.config?.tags.map((tag) => <TagColored key={tag} value={tag}/>)
                ) : (
                  <b>-</b>
                )}
              </div>

              <Service
                key={serviceName}
                proxyName={instance.proxyName}
                instanceName={instance.name || ""}
                serviceName={serviceName}
                service={service}
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
      <div style={{ display: "flex", gap: "1rem", flexWrap: "wrap" }}>
        {managerStore.instances.map((instance) => {
          return <Instance key={`${instance.proxyName}_${instance.name}`} instance={instance} />;
        })}
      </div>

      <DebugJSON data={managerStore.instances} title="Instances" />
    </div>
  );
});

export default InstanceList;
