import { useEffect } from "react";

import { useParams } from "react-router";

import { Spin, Typography } from "antd";

import { DebugJSON } from "components";

import { observer, useStores } from "stores";

const Stats = observer(() => {
  const { managerStore } = useStores();
  const params = useParams<{
		proxyId?:string;
    instanceId: string;
    serviceId: string;
    taskId: string;
    destinationId: string;
  }>();

  useEffect(() => {
    if (params.instanceId && params.serviceId && params.taskId && params.destinationId) {
      managerStore.fetchStats(
        params.instanceId,
        params.serviceId,
        params.taskId,
        params.destinationId,
				params.proxyId
      );
    }

    return () => {
      managerStore.resetStats();
    };
  }, [params]);

  return (
    <div>
      <Typography.Title>Stats</Typography.Title>
      <Typography.Title level={3}>
        {[params.instanceId, params.serviceId, params.taskId, params.destinationId]
          .filter(Boolean)
          .join(" / ")}
      </Typography.Title>

      {managerStore.statsStatus?.state === "pending" ? (
        <Spin size="large" />
      ) : (
        <>
          <p>Total Size: {managerStore.stats?.data?.totalSizeFormatted}</p>

          <p>Total File Count: {managerStore.stats?.data?.totalFileCount}</p>

          <p>SnapshotsCount: {managerStore.stats?.data?.snapshotsCount}</p>
        </>
      )}

      <br />

      <DebugJSON data={managerStore.stats?.data} />
    </div>
  );
});

export default Stats;
