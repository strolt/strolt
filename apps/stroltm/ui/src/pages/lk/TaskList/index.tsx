import { useMemo } from "react";

import type { ColumnsType } from "antd/es/table";
import type { CompareFn } from "antd/es/table/interface";
import type { TaskListItem } from "stores/manager.store/taskList";

import { Print, Table } from "components";
import { observer } from "mobx-react-lite";
import { managerStore } from "stores/manager.store";
import { getTagKey } from "utils";

import { BackupAllButton } from "./components/BackupAllButton";
import { BackupButton } from "./components/BackupButton";
import { PrintDestinations } from "./components/PrintDestinations";
import { PrintNotifications } from "./components/PrintNotifications";
import { PrintSchedule } from "./components/PrintSchedule";
import { PrintUptime } from "./components/PrintUptime";
import { PrintVersion } from "./components/PrintVersion";

const nameSorter = (field: keyof TaskListItem): CompareFn<TaskListItem> => {
  return (a, b, order) => {
    const _a = (!!a?.[field] ? String(a?.[field]) : "").toLowerCase();
    const _b = (!!b?.[field] ? String(b?.[field]) : "").toLowerCase();

    if (_a < _b) {
      return -1;
    }
    if (_a > _b) {
      return 1;
    }

    return 0;
  };
};

const deleteDuplicates = (list: string[]) => {
  return list.reduce<string[]>((acc, el) => {
    if (!acc.includes(el)) {
      acc.push(el);
    }
    return acc;
  }, []);
};

const useColumns = (list: TaskListItem[]): ColumnsType<TaskListItem> => {
  return useMemo(() => {
    const columns: ColumnsType<TaskListItem> = [
      {
        key: "backupButton",
        render: (_, r) => (
          <BackupButton
            instanceName={r.instanceName}
            isDisabled={!r.isOnline}
            proxyName={r.proxyName}
            serviceName={r.serviceName}
            taskName={r.taskName}
          />
        ),
        width: "7rem",
      },
      {
        dataIndex: "isOnline",
        defaultSortOrder: "ascend",
        key: "isOnline",
        render: (v) => (
          <div style={{ display: "flex", justifyContent: "center" }}>
            <Print.Boolean size="1.2rem" value={v} />
          </div>
        ),
        sorter: {
          compare: (a, b) => {
            if (+a.isOnline < +b.isOnline) {
              return -1;
            }

            if (+a.isOnline > +b.isOnline) {
              return 1;
            }
            return 0;
          },
          multiple: 1,
        },
        title: "online",
        width: "1rem",
      },
      {
        dataIndex: "proxyName",
        defaultSortOrder: "ascend",
        filterResetToDefaultFilteredValue: true,
        filters: deleteDuplicates(list.map((el) => el.proxyName || ""))
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        key: "proxyName",
        onFilter: (value, r) => r.proxyName === value,
        render: (v) => <Print.Text value={v} />,
        sorter: { compare: nameSorter("proxyName") },
        title: "proxy",
      },
      {
        dataIndex: "instanceName",
        defaultSortOrder: "ascend",
        filterResetToDefaultFilteredValue: true,
        filters: deleteDuplicates(list.map((el) => el.instanceName || ""))
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        key: "instanceName",
        onFilter: (value, r) => r.instanceName === value,
        render: (v) => <Print.Text value={v} />,
        sorter: { compare: nameSorter("instanceName") },
        title: "instance",
      },
      {
        dataIndex: "serviceName",
        defaultSortOrder: "ascend",
        filterResetToDefaultFilteredValue: true,
        filters: deleteDuplicates(list.map((el) => el.serviceName || ""))
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        key: "serviceName",
        onFilter: (value, r) => r.serviceName === value,
        render: (v) => <Print.Text value={v} />,
        sorter: { compare: nameSorter("serviceName") },
        title: "service",
      },
      {
        dataIndex: "taskName",
        defaultSortOrder: "ascend",
        filterResetToDefaultFilteredValue: true,
        filters: deleteDuplicates(list.map((el) => el.taskName || ""))
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        key: "taskName",
        onFilter: (value, r) => r.taskName === value,
        render: (v) => <Print.Text value={v} />,
        sorter: { compare: nameSorter("taskName") },
        title: "task",
      },
      {
        dataIndex: "timezone",
        filterResetToDefaultFilteredValue: true,
        filters: deleteDuplicates(list.map((el) => el.timezone)).map((v) => ({
          text: v,
          value: v,
        })),
        key: "timezone",
        onFilter: (value, r) => r.timezone === String(value),
        sorter: { compare: nameSorter("taskName") },
        title: "timezone",
      },
      {
        dataIndex: "tags",
        filterResetToDefaultFilteredValue: true,
        filters: deleteDuplicates(
          list
            .map((el) => el.tags)
            .flat()
            .map((el) => getTagKey(el)),
        ).map((v) => ({
          text: v,
          value: v,
        })),
        key: "tags",
        onCell: () => ({ style: { maxWidth: "25rem" } }),
        onFilter: (value, r) => !!r.tags.find((tag) => tag.startsWith(String(value))),
        render: (v) => <Print.TagList fallback={<>-</>} value={v} />,
        title: "tags",
      },
      {
        dataIndex: "schedule",
        key: "schedule",
        render: (_, r) => <PrintSchedule {...r.schedule} />,
        title: "schedule",
      },
      {
        dataIndex: "source",
        filterResetToDefaultFilteredValue: true,
        filters: deleteDuplicates(list.map((el) => el.source.driver))
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        key: "source",
        onFilter: (value, r) => r.source.driver === String(value),
        render: (_, r) => <Print.Text value={r.source.driver} />,
        sorter: { compare: nameSorter("source") },
        title: "source",
      },
      {
        dataIndex: "destinations",
        filterResetToDefaultFilteredValue: true,
        filters: deleteDuplicates(list.map((el) => el.destinations.map((el) => el.driver)).flat())
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        key: "destinations",
        onFilter: (value, r) => !!r.destinations.find((el) => el.driver === String(value)),
        render: (_, r) => (
          <PrintDestinations
            list={r.destinations.map((el) => ({
              ...el,
              destinationName: el.name,
              instanceName: r.instanceName,
              proxyName: r.proxyName,
              serviceName: r.serviceName,
              taskName: r.taskName,
            }))}
          />
        ),
        title: "destinations",
      },
      {
        dataIndex: "notifications",
        filterResetToDefaultFilteredValue: true,
        filters: deleteDuplicates(list.map((el) => el.notifications.map((el) => el.driver)).flat())
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        key: "notifications",
        onCell: () => ({ style: { maxWidth: "20rem" } }),
        onFilter: (value, r) => !!r.notifications.find((el) => el.driver === String(value)),
        render: (_, r) => <PrintNotifications list={r.notifications} />,
        title: "notifications",
      },
      {
        key: "version",
        render: (_, r) => <PrintVersion instanceName={r.instanceName} proxyName={r.proxyName} />,
        title: "version",
      },
      {
        dataIndex: "uptime",
        key: "uptime",
        render: (v) => <PrintUptime uptime={v} />,
        sorter: { compare: (a, b) => a.uptime - b.uptime },
        title: "uptime",
      },
    ];

    columns.map((el, i, list) => {
      const multiple = list.length - i;
      if (!el.sorter) {
        return el;
      }

      if (typeof el.sorter === "object") {
        el.sorter.multiple = multiple;
      }

      return el;
    });

    return columns;
  }, [list]);
};

const TaskList = observer(() => {
  const columns = useColumns(managerStore.taskList);

  const count = useMemo(() => {
    const proxyInstances = managerStore.instances
      .map((el) => el.proxyName)
      .reduce<string[]>((acc, el) => {
        if (el && !acc.includes(el)) {
          acc.push(el);
        }
        return acc;
      }, []).length;
    let stroltInstancesDirect = 0;
    let stroltInstancesProxy = 0;

    managerStore.instances.forEach((el) => {
      if (el.proxyName) {
        stroltInstancesProxy++;
      } else {
        stroltInstancesDirect++;
      }
    });

    return {
      proxyInstances,
      stroltInstancesDirect,
      stroltInstancesProxy,
      tasks: managerStore.taskList.length,
    };
  }, [managerStore.instances, managerStore.taskList]);

  return (
    <>
      <BackupAllButton />
      <Table
        columns={columns}
        dataSource={managerStore.taskList}
        footer={() => (
          <>
            <div>
              Tasks: <b>{count.tasks}</b>
            </div>
            <div>
              Proxy instances: <b>{count.proxyInstances}</b>
            </div>
            <div>
              Strolt instances (direct): <b>{count.stroltInstancesDirect}</b>
            </div>
            <div>
              Strolt instances (proxy): <b>{count.stroltInstancesProxy}</b>
            </div>
          </>
        )}
        pagination={false}
        rowKey="key"
        scroll={{ x: "max-content" }}
      />
      {/* <ReactJson src={managerStore.taskList || {}} /> */}
    </>
  );
});

export default TaskList;
