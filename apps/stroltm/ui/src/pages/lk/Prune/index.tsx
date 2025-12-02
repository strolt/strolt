import { FC, useEffect, useState } from "react";

import { useParams } from "react-router";

import { Button, Popconfirm, Table, Tag, Typography } from "antd";
import { ColumnsType } from "antd/es/table";

import { DebugJSON, Print, TagColored } from "components";

import { Snapshot } from "api/generated";

import { observer, useStores } from "stores";

interface PruneButtonProps {
  instanceId?: string;
  serviceId?: string;
  taskId?: string;
  destinationId?: string;
  count?: number;
  proxyId?: string;
}
const PruneButton: FC<PruneButtonProps> = observer(
  ({ proxyId, instanceId, serviceId, taskId, destinationId, count }) => {
    const { managerStore } = useStores();

    const handleClick = async () => {
      if (instanceId && serviceId && taskId && destinationId) {
        await managerStore.fetchPrune(instanceId, serviceId, taskId, destinationId, proxyId);

        managerStore.fetchSnapshotsForPrune(instanceId, serviceId, taskId, destinationId, proxyId);
      }
    };

    return (
      <Popconfirm title="Are you sure?" onConfirm={handleClick} okText="Yes">
        <Button
          disabled={!count}
          type="primary"
          danger
          loading={managerStore.pruneStatus?.state === "pending"}
        >
          Prune{!!count && ` (${count})`}
        </Button>
      </Popconfirm>
    );
  },
);

const columns: ColumnsType<Snapshot> = [
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
          <TagColored key={tag} value={tag} />
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

const Prune = observer(() => {
  const { managerStore } = useStores();
  const params = useParams<{
    proxyId?: string;
    instanceId: string;
    serviceId: string;
    taskId: string;
    destinationId: string;
  }>();

  const [expandedKey, setExpandedKey] = useState("");

  useEffect(() => {
    if (params.instanceId && params.serviceId && params.taskId && params.destinationId) {
      managerStore.fetchSnapshotsForPrune(
        params.instanceId,
        params.serviceId,
        params.taskId,
        params.destinationId,
        params.proxyId,
      );
    }

    return () => {
      managerStore.resetSnapshotsForPrune();
    };
  }, [params]);

  return (
    <div>
      <Typography.Title>Snapshot List For Prune</Typography.Title>
      <Typography.Title level={3}>
        {[params.instanceId, params.serviceId, params.taskId, params.destinationId]
          .filter(Boolean)
          .join(" / ")}
      </Typography.Title>

      <PruneButton
        proxyId={params.proxyId}
        instanceId={params.instanceId}
        serviceId={params.serviceId}
        taskId={params.taskId}
        destinationId={params.destinationId}
        count={managerStore.snapshotsForPrune?.data?.length}
      />

      <Table
        dataSource={managerStore.snapshotsForPrune?.data}
        columns={columns}
        loading={managerStore.snapshotsForPruneStatus?.state === "pending"}
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
        footer={() => <b>Total: {managerStore.snapshotsForPrune?.data?.length || 0}</b>}
      />

      <br />

      <DebugJSON data={managerStore.snapshotsForPrune?.data} />
    </div>
  );
});

export default Prune;
