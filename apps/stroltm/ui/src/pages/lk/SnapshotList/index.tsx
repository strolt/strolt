import { useEffect, useState } from "react";

import { useParams } from "react-router";

import { Table, Tag, Typography } from "antd";
import { ColumnsType } from "antd/es/table";

import { DebugJSON, Print } from "components";

import { ModelsInterfacesSnapshot } from "api/generated";

import { observer, useStores } from "stores";

const columns: ColumnsType<ModelsInterfacesSnapshot> = [
  {
    title: "Short ID",
    dataIndex: "shortId",
    key: "shortId",
  },
  {
    title: "ID",
    dataIndex: "id",
    key: "id",
  },
  {
    title: "Tags",
    dataIndex: "tags",
    key: "tags",
    render: (tags: string[]) => (
      <>
        {tags.map((tag) => (
          <Tag key={tag}>{tag}</Tag>
        ))}
      </>
    ),
  },
  {
    title: "Time",
    dataIndex: "time",
    key: "time",
    render: (v) => <Print.Time value={v} withTime />,
  },
];

const SnapshotList = observer(() => {
  const { managerStore } = useStores();
  const params = useParams<{
    instanceId: string;
    serviceId: string;
    taskId: string;
    destinationId: string;
  }>();

  const [expandedKey, setExpandedKey] = useState("");

  useEffect(() => {
    if (params.instanceId && params.serviceId && params.taskId && params.destinationId) {
      managerStore.fetchSnapshots(
        params.instanceId,
        params.serviceId,
        params.taskId,
        params.destinationId,
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
        {[params.instanceId, params.serviceId, params.taskId, params.destinationId]
          .filter(Boolean)
          .join(" / ")}
      </Typography.Title>

      <Table
        dataSource={managerStore.snapshots?.items}
        columns={columns}
        loading={managerStore.snapshotsStatus?.state === "pending"}
        rowKey="id"
        pagination={false}
        scroll={{
          x: "max-content",
        }}
        expandable={{
          expandedRowKeys: expandedKey ? [expandedKey] : [],
          expandedRowRender: (data) => (
            <>
              <b>paths:</b>
              <ul>
                {data.paths?.map((path) => (
                  <li key={path}>{path}</li>
                ))}
              </ul>
            </>
          ),
          onExpand: (expanded, record) => {
            if (expanded && record.id) {
              setExpandedKey(record.id);
            } else {
              setExpandedKey("");
            }
          },
        }}
        footer={() => <b>Total: {managerStore.snapshots?.items?.length || 0}</b>}
      />

      <br />

      <DebugJSON data={managerStore.snapshots || {}} />
    </div>
  );
});

export default SnapshotList;
