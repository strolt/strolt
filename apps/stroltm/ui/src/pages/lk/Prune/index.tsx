import type { FC } from "react";
import { useEffect, useState } from "react";
import { useParams } from "react-router";

import { Button, Popconfirm, Table, Typography } from "antd";

import type { ColumnsType } from "antd/es/table";
import type { Snapshot } from "api/generated";

import { DebugJSON, Print, TagColored } from "components";
import { observer, useStores } from "stores";

interface PruneButtonProps {
  count?: number;
  destinationId?: string;
  instanceId?: string;
  proxyId?: string;
  serviceId?: string;
  taskId?: string;
}
const PruneButton: FC<PruneButtonProps> = observer(
  ({ count, destinationId, instanceId, proxyId, serviceId, taskId }) => {
    const { managerStore } = useStores();

    const handleClick = async () => {
      if (instanceId && serviceId && taskId && destinationId) {
        await managerStore.fetchPrune(instanceId, serviceId, taskId, destinationId, proxyId);

        managerStore.fetchSnapshotsForPrune(instanceId, serviceId, taskId, destinationId, proxyId);
      }
    };

    return (
      <Popconfirm okText="Yes" onConfirm={handleClick} title="Are you sure?">
        <Button
          danger
          disabled={!count}
          loading={managerStore.pruneStatus?.state === "pending"}
          type="primary"
        >
          Prune{Boolean(count) && ` (${count})`}
        </Button>
      </Popconfirm>
    );
  },
);

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

const Prune = observer(() => {
  const { managerStore } = useStores();
  const params = useParams<{
    destinationId: string;
    instanceId: string;
    proxyId?: string;
    serviceId: string;
    taskId: string;
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
        count={managerStore.snapshotsForPrune?.data?.length}
        destinationId={params.destinationId}
        instanceId={params.instanceId}
        proxyId={params.proxyId}
        serviceId={params.serviceId}
        taskId={params.taskId}
      />

      <Table
        columns={columns}
        dataSource={managerStore.snapshotsForPrune?.data}
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
        footer={() => <b>Total: {managerStore.snapshotsForPrune?.data?.length || 0}</b>}
        loading={managerStore.snapshotsForPruneStatus?.state === "pending"}
        pagination={false}
        rowKey="id"
        scroll={{
          x: "max-content",
        }}
      />

      <br />

      <DebugJSON data={managerStore.snapshotsForPrune?.data} />
    </div>
  );
});

export default Prune;
