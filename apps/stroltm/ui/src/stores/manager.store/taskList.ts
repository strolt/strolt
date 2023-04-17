import { infoStore } from "stores/info.store";

import * as apiGenerated from "../../api/generated";

export interface TaskListItemSource {
  driver: string;
}

export interface TaskListItemDestination {
  name: string;
  driver: string;
}

export interface TaskListItemNotification {
  name: string;
  driver: string;
  events: string[];
}

export interface TaskListItemSchedule {
  backup: string;
  prune: string;
}

export interface TaskListItem {
  key: string;
  proxyName?: string;
  instanceName?: string;
  serviceName?: string;
  taskName?: string;
  isOnline: boolean;
  workJobs: number;
  tags: string[];
  timezone: string;
  source: TaskListItemSource;
  destinations: TaskListItemDestination[];
  notifications: TaskListItemNotification[];
  schedule: TaskListItemSchedule;
  uptime: number;
}

const createTask = (instance: apiGenerated.ManagerPreparedInstance): TaskListItem => {
  const isOnline =
    !!instance.name &&
    !!infoStore.map.get(infoStore.getKey(instance.name, instance.proxyName))?.isOnline;

  return {
    key: "",
    proxyName: instance.proxyName,
    instanceName: instance.name,
    isOnline: isOnline,
    workJobs: instance.taskStatus?.tasks?.length || 0,
    tags: [],
    timezone: "UTC",
    source: { driver: "" },
    destinations: [],
    notifications: [],
    schedule: {
      backup: "",
      prune: "",
    },
    uptime: 0,
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
              name: notification.name || "",
              events: notification.events || [],
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
      } catch (e) {
        console.log(e);
      }
    }

    if (!instance.isOnline && instance.lastestOnlineAt) {
      try {
        const date = new Date(instance.lastestOnlineAt);
        if (date.getTime() < 0) {
          return ms;
        }
        ms = (Date.now() - date.getTime()) * -1;
      } catch (e) {
        console.log(e);
      }
    }
  }

  return ms;
};
