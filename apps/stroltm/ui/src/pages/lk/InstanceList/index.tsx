import { FC, useEffect } from "react";

import { Button, Card, message, Popconfirm, Tag, Typography } from "antd";

import { DebugJSON, Link } from "components";

import { ManagerhGetInstancesResultItem, ModelsAPIConfigServiceTask } from "api/generated";

import { observer, useStores } from "stores";

export interface BackupButtonProps {
  instanceName: string;
  serviceName: string;
  taskName: string;
}
const BackupButton: FC<BackupButtonProps> = observer(({ instanceName, serviceName, taskName }) => {
  const { managerStore } = useStores();

  const status = managerStore.backupStatusMap.get(
    managerStore.backupStatusMapKey(instanceName, serviceName, taskName),
  );

  return (
    <Popconfirm
      title="Are you sure?"
      onConfirm={() => managerStore.backup(instanceName, serviceName, taskName)}
    >
      <Button loading={status?.state === "pending"} size="small" danger>
        Backup
      </Button>
    </Popconfirm>
  );
});

const BackupAll: FC = observer(() => {
  const { managerStore } = useStores();

  useEffect(() => {
    return () => managerStore.resetBackupAll();
  }, []);

  const handleClick = async () => {
    const data = await managerStore.backupAll();
    message.info(`Success started: ${data.successStarted?.length}`);
    message.error(`Error started: ${data.errorStarted?.length}`);
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
  instanceName: string;
  serviceName: string;
  taskName: string;
  task: ModelsAPIConfigServiceTask;
}
const Task: FC<TaskProps> = observer(({ task, instanceName, serviceName, taskName }) => {
  return (
    <Card
      size="small"
      style={{ marginBottom: "1rem" }}
      title={`task: [${taskName}]`}
      extra={
        <BackupButton instanceName={instanceName} serviceName={serviceName} taskName={taskName} />
      }
    >
      <div>
        TAGS:
        {task?.tags?.length ? task?.tags.map((tag) => <Tag key={tag}>{tag}</Tag>) : <b>-</b>}
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
                  to="instances.instanceId.serviceId.taskId.destinationId.snapshotList"
                  params={{
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
                  <Tag key={eventName}>{eventName}</Tag>
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
  serviceName: string;
  service: Record<string, ModelsAPIConfigServiceTask>;
}
const Service: FC<ServiceProps> = observer(({ service, serviceName, instanceName }) => {
  return (
    <Card size="small" style={{ marginBottom: "1rem" }} title={`service: [${serviceName}]`}>
      {Object.entries(service).map(([taskName, task]) => {
        return (
          <Task
            key={taskName}
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
  instance: ManagerhGetInstancesResultItem;
}
const Instance: FC<InstanceProps> = observer(({ instance }) => {
  return (
    <div style={{ minWidth: "25rem" }}>
      <Card
        size="small"
        title={`instance: [${instance.instanceName}] (${instance.config?.timezone})`}
        extra={
          instance.isOnline ? <Tag color="success">Online</Tag> : <Tag color="error">Offline</Tag>
        }
      >
        {Object.entries(instance.config?.services || {}).map(([serviceName, service]) => {
          return (
            <>
              <div>
                TAGS:
                {instance.config?.tags?.length ? (
                  instance.config?.tags.map((tag) => <Tag key={tag}>{tag}</Tag>)
                ) : (
                  <b>-</b>
                )}
              </div>

              <Service
                key={serviceName}
                instanceName={instance.instanceName || ""}
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

  useEffect(() => {
    managerStore.setIsAutoUpdateInstancesEnabled(true);
    return () => {
      managerStore.setIsAutoUpdateInstancesEnabled(false);
      // managerStore.resetBackup();
      // managerStore.resetInstances();
    };
  }, []);

  return (
    <div>
      <Typography.Title>Instances:</Typography.Title>
      <BackupAll />
      <div style={{ display: "flex", gap: "1rem", flexWrap: "wrap" }}>
        {managerStore.instances.map((instance) => {
          return <Instance key={instance.instanceName} instance={instance} />;
        })}
      </div>

      <DebugJSON data={managerStore.instances} title="Instances" />
    </div>
  );
});

export default InstanceList;
