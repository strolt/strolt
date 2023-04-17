import { observer } from "mobx-react-lite";

import { useMemo } from "react";

import { ColumnsType } from "antd/es/table";
import { CompareFn } from "antd/es/table/interface";

import { BackupAllButton } from "./components/BackupAllButton";
import { BackupButton } from "./components/BackupButton";
import { PrintDestinations } from "./components/PrintDestinations";
import { PrintNotifications } from "./components/PrintNotifications";
import { PrintSchedule } from "./components/PrintSchedule";
import { PrintUptime } from "./components/PrintUptime";
import { PrintVersion } from "./components/PrintVersion";
import { Print, Table } from "components";

import { managerStore } from "stores/manager.store";
import { TaskListItem } from "stores/manager.store/taskList";

import { getTagKey } from "utils";

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
            proxyName={r.proxyName}
            instanceName={r.instanceName}
            serviceName={r.serviceName}
            taskName={r.taskName}
            isDisabled={!r.isOnline}
          />
        ),
        width: "7rem",
      },
      {
        title: "online",
        dataIndex: "isOnline",
        key: "isOnline",
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
        render: (v) => (
          <div style={{ display: "flex", justifyContent: "center" }}>
            <Print.Boolean size="1.2rem" value={v} />
          </div>
        ),
        defaultSortOrder: "ascend",
        width: "1rem",
      },
      {
        title: "proxy",
        dataIndex: "proxyName",
        key: "proxyName",
        filters: deleteDuplicates(list.map((el) => el.proxyName || ""))
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        onFilter: (value, r) => r.proxyName === value,
        filterResetToDefaultFilteredValue: true,
        sorter: { compare: nameSorter("proxyName") },
        render: (v) => <Print.Text value={v} />,
        defaultSortOrder: "ascend",
      },
      {
        title: "instance",
        dataIndex: "instanceName",
        key: "instanceName",
        filters: deleteDuplicates(list.map((el) => el.instanceName || ""))
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        onFilter: (value, r) => r.instanceName === value,
        filterResetToDefaultFilteredValue: true,
        sorter: { compare: nameSorter("instanceName") },
        render: (v) => <Print.Text value={v} />,
        defaultSortOrder: "ascend",
      },
      {
        title: "service",
        dataIndex: "serviceName",
        key: "serviceName",
        filters: deleteDuplicates(list.map((el) => el.serviceName || ""))
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        onFilter: (value, r) => r.serviceName === value,
        filterResetToDefaultFilteredValue: true,
        sorter: { compare: nameSorter("serviceName") },
        render: (v) => <Print.Text value={v} />,
        defaultSortOrder: "ascend",
      },
      {
        title: "task",
        dataIndex: "taskName",
        key: "taskName",
        filters: deleteDuplicates(list.map((el) => el.taskName || ""))
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        onFilter: (value, r) => r.taskName === value,
        filterResetToDefaultFilteredValue: true,
        sorter: { compare: nameSorter("taskName") },
        render: (v) => <Print.Text value={v} />,
        defaultSortOrder: "ascend",
      },
      {
        title: "timezone",
        dataIndex: "timezone",
        key: "timezone",
        sorter: { compare: nameSorter("taskName") },
        filters: deleteDuplicates(list.map((el) => el.timezone)).map((v) => ({
          text: v,
          value: v,
        })),
        onFilter: (value, r) => r.timezone === String(value),
        filterResetToDefaultFilteredValue: true,
      },
      {
        title: "tags",
        dataIndex: "tags",
        key: "tags",
        render: (v) => <Print.TagList value={v} />,
        filters: deleteDuplicates(
          list
            .map((el) => el.tags)
            .flat()
            .map((el) => getTagKey(el)),
        ).map((v) => ({
          text: v,
          value: v,
        })),
        onFilter: (value, r) => !!r.tags.find((tag) => tag.startsWith(String(value))),
        filterResetToDefaultFilteredValue: true,
        onCell: () => ({ style: { maxWidth: "25rem" } }),
      },
      {
        title: "schedule",
        dataIndex: "schedule",
        key: "schedule",
        render: (_, r) => <PrintSchedule {...r.schedule} />,
      },
      {
        title: "source",
        dataIndex: "source",
        key: "source",
        sorter: { compare: nameSorter("source") },
        filters: deleteDuplicates(list.map((el) => el.source.driver))
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        onFilter: (value, r) => r.source.driver === String(value),
        filterResetToDefaultFilteredValue: true,
        render: (_, r) => <Print.Text value={r.source.driver} />,
      },
      {
        title: "destinations",
        dataIndex: "destinations",
        key: "destinations",
        filters: deleteDuplicates(list.map((el) => el.destinations.map((el) => el.driver)).flat())
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        onFilter: (value, r) => !!r.destinations.find((el) => el.driver === String(value)),
        filterResetToDefaultFilteredValue: true,
        render: (_, r) => (
          <PrintDestinations
            list={r.destinations.map((el) => ({
              ...el,
              proxyName: r.proxyName,
              instanceName: r.instanceName,
              serviceName: r.serviceName,
              taskName: r.taskName,
              destinationName: el.name,
            }))}
          />
        ),
      },
      {
        title: "notifications",
        dataIndex: "notifications",
        key: "notifications",
        filters: deleteDuplicates(list.map((el) => el.notifications.map((el) => el.driver)).flat())
          .filter(Boolean)
          .map((v) => ({ text: v, value: v })),
        onFilter: (value, r) => !!r.notifications.find((el) => el.driver === String(value)),
        filterResetToDefaultFilteredValue: true,
        render: (_, r) => <PrintNotifications list={r.notifications} />,
        onCell: () => ({ style: { maxWidth: "20rem" } }),
      },
      {
        title: "version",
        key: "version",
        render: (_, r) => <PrintVersion proxyName={r.proxyName} instanceName={r.instanceName} />,
      },
      {
        title: "uptime",
        dataIndex: "uptime",
        key: "uptime",
        render: (v) => <PrintUptime uptime={v} />,
        sorter: { compare: (a, b) => a.uptime - b.uptime },
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
      stroltInstancesDirect,
      stroltInstancesProxy,
      proxyInstances,
      tasks: managerStore.taskList.length,
    };
  }, [managerStore.instances, managerStore.taskList]);

  return (
    <>
      <BackupAllButton />
      <Table
        scroll={{ x: "max-content" }}
        dataSource={managerStore.taskList}
        rowKey="key"
        columns={columns}
        pagination={false}
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
      />
      {/* <ReactJson src={managerStore.taskList || {}} /> */}
    </>
  );
});

export default TaskList;
