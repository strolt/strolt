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
}

const createTask = (instance: apiGenerated.ManagerhManagerPreparedInstance): TaskListItem => {
  return {
    key: "",
    proxyName: instance.proxyName,
    instanceName: instance.name,
    isOnline: !!instance.isOnline,
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
  };
};

export const getTaskList = (
  instanceList: apiGenerated.ManagerhManagerPreparedInstance[],
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
  }));
};