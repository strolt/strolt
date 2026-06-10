import { useEffect, useState } from "react";
import { useParams } from "react-router";

import { Table, Typography } from "antd";

import type { ColumnsType } from "antd/es/table";
import type { Snapshot } from "api/generated";

import { DebugJSON, Print, TagColored } from "components";
import { observer, useStores } from "stores";

const columns: ColumnsType<Snapshot> = [
  {
    dataIndex: "shortId",
    key: "shortId",
    title: "Short ID",
  },
  {
    dataIndex: "id",
    key: "id",
    title: "ID",
  },
  {
    dataIndex: "tags",
    key: "tags",
    render: (tags: string[]) => (
      <>
        {tags.map((tag) => (
          <TagColored key={tag} value={tag} />
        ))}
      </>
    ),
    title: "Tags",
  },
  {
    dataIndex: "time",
    key: "time",
    render: (v) => <Print.Time value={v} withTime />,
    title: "Time",
  },
];

const SnapshotList = observer(() => {
  const { managerStore } = useStores();
  const params = useParams<{
    destinationId: string;
    instanceId: string;
    proxyId: string;
    serviceId: string;
    taskId: string;
  }>();

  const [expandedKey, setExpandedKey] = useState("");

  useEffect(() => {
    if (params.instanceId && params.serviceId && params.taskId && params.destinationId) {
      managerStore.fetchSnapshots(
        params.instanceId,
        params.serviceId,
        params.taskId,
        params.destinationId,
        params.proxyId,
      );
    }

    return () => {
      managerStore.resetSnapshots();
    };
  }, [params]);

  return (
    <div>
      <Typography.Title>Snapshot List</Typography.Title>
      <Typography.Title level={3}>
        {[params.proxyId, params.instanceId, params.serviceId, params.taskId, params.destinationId]
          .filter(Boolean)
          .join(" / ")}
      </Typography.Title>

      <Table
        columns={columns}
        dataSource={managerStore.snapshots?.items}
        expandable={{
          expandedRowKeys: expandedKey ? [expandedKey] : [],
          expandedRowRender: (data: Snapshot) => (
            <>
              <b>paths:</b>
              <ul>
                {data.paths?.map((path: string) => (
                  <li key={path}>{path}</li>
                ))}
              </ul>
            </>
          ),
          onExpand: (expanded: boolean, record: Snapshot) => {
            if (expanded && record.id) {
              setExpandedKey(record.id);
            } else {
              setExpandedKey("");
            }
          },
        }}
        footer={() => <b>Total: {managerStore.snapshots?.items?.length || 0}</b>}
        loading={managerStore.snapshotsStatus?.state === "pending"}
        pagination={false}
        rowKey="id"
        scroll={{
          x: "max-content",
        }}
      />

      <br />

      <DebugJSON data={managerStore.snapshots || {}} />
    </div>
  );
});

export default SnapshotList;
