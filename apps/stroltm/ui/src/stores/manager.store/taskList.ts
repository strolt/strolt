import { infoStore } from "stores/info.store";

import type * as apiGenerated from "../../api/generated";

export interface TaskListItem {
  destinations: TaskListItemDestination[];
  instanceName?: string;
  isOnline: boolean;
  key: string;
  notifications: TaskListItemNotification[];
  proxyName?: string;
  schedule: TaskListItemSchedule;
  serviceName?: string;
  source: TaskListItemSource;
  tags: string[];
  taskName?: string;
  timezone: string;
  uptime: number;
  workJobs: number;
}

export interface TaskListItemDestination {
  driver: string;
  name: string;
}

export interface TaskListItemNotification {
  driver: string;
  events: string[];
  name: string;
}

export interface TaskListItemSchedule {
  backup: string;
  prune: string;
}

export interface TaskListItemSource {
  driver: string;
}

const createTask = (instance: apiGenerated.ManagerPreparedInstance): TaskListItem => {
  const isOnline =
    instance.name !== undefined &&
    Boolean(infoStore.map.get(infoStore.getKey(instance.name, instance.proxyName))?.isOnline);

  return {
    destinations: [],
    instanceName: instance.name,
    isOnline: isOnline,
    key: "",
    notifications: [],
    proxyName: instance.proxyName,
    schedule: {
      backup: "",
      prune: "",
    },
    source: { driver: "" },
    tags: [],
    timezone: "UTC",
    uptime: 0,
    workJobs: instance.taskStatus?.tasks?.length || 0,
  };
};

export const getTaskList = (
  instanceList: apiGenerated.ManagerPreparedInstance[],
): TaskListItem[] => {
  const list: TaskListItem[] = [];

  instanceList.forEach((instance) => {
    if (!instance.config?.services) {
      const task = createTask(instance);

      list.push(task);
    }

    Object.entries(instance.config?.services || {}).forEach(([serviceName, service]) => {
      Object.entries(service || {}).forEach(([taskName, taskItem]) => {
        const task = createTask(instance);
        task.serviceName = serviceName;
        task.taskName = taskName;
        task.tags = (instance.config?.tags || []).concat(taskItem.tags || []);
        task.source.driver = taskItem.source?.driver || "";
        if (instance.config?.timezone) {
          task.timezone = instance.config.timezone;
        }
        task.destinations = Object.entries(taskItem.destinations || {}).reduce<
          TaskListItemDestination[]
        >((acc, [destinationName, destination]) => {
          acc.push({ driver: destination.driver || "", name: destinationName });
          return acc;
        }, []);
        task.notifications = (taskItem.notifications || []).reduce<TaskListItemNotification[]>(
          (acc, notification) => {
            acc.push({
              driver: notification.driver || "",
              events: notification.events || [],
              name: notification.name || "",
            });
            return acc;
          },
          [],
        );
        task.schedule.backup = taskItem.schedule?.backup || "";
        task.schedule.prune = taskItem.schedule?.prune || "";
        list.push(task);
      });
    });
  });

  return list.map((el) => ({
    ...el,
    key: [el.proxyName, el.instanceName, el.serviceName, el.taskName].join("_"),
    uptime: getUptime(el.proxyName, el.instanceName),
  }));
};

const getUptime = (proxyName?: string, instanceName?: string) => {
  let ms = 0;

  if (instanceName) {
    const instance = infoStore.map.get(infoStore.getKey(instanceName, proxyName));

    if (!instance) {
      return ms;
    }

    if (instance.isOnline && instance.startedAt) {
      try {
        const date = new Date(instance.startedAt);
        ms = Date.now() - date.getTime();
      } catch (error) {
        console.log(error);
      }
    }

    if (!instance.isOnline && instance.lastestOnlineAt) {
      try {
        const date = new Date(instance.lastestOnlineAt);
        if (date.getTime() < 0) {
          return ms;
        }
        ms = (Date.now() - date.getTime()) * -1;
      } catch (error) {
        console.log(error);
      }
    }
  }

  return ms;
};
