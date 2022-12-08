import { useEffect } from "react";
import { observer, useStores } from "./stores";

const App = observer(() => {
	
	const { managerStore, authStore } = useStores();

	// useEffect(() => {
	// 	managerStore.fetchInstances();
	// }, []);

	if (!authStore.isAuthorized){
		return <div>
			NOT AUTHORIZED
			<button onClick={()=>authStore.setCredentials("admin","admin")}>Login</button>
		</div>
	}

	return (
		<div>
			{managerStore?.snapshots && (
				<div
					style={{
						border: "1px solid #ccc",
						marginBottom: "1rem",
						padding: "0.5rem",
					}}
				>
					<div>
						SNAPSHOTS:{" "}
						<button onClick={() => managerStore.resetSnapshots()}>RESET</button>
					</div>
					<div>lastUpdated: {managerStore?.snapshots?.lastUpdated}</div>
					{managerStore?.snapshots?.data?.map((snapshot) => (
						<div key={snapshot.id}>
							<ul>
								<li>id: {snapshot.id}</li>
								<li>shortId: {snapshot.shortId}</li>
								<li>paths: [{snapshot.paths?.join(", ")}]</li>
								<li>tags: [{snapshot.tags?.join(", ")}]</li>
								<li>time: {snapshot.time}</li>
							</ul>
						</div>
					))}
				</div>
			)}

			{managerStore?.snapshotsForPrune && (
				<div
					style={{
						border: "1px solid #ccc",
						marginBottom: "1rem",
						padding: "0.5rem",
					}}
				>
					<div>
						SNAPSHOTS FOR PRUNE:{" "}
						<button onClick={() => managerStore.resetSnapshotsForPrune()}>
							RESET
						</button>
					</div>
					<pre>
						{JSON.stringify(managerStore?.snapshotsForPrune || {}, null, 2)}
					</pre>
				</div>
			)}

			{managerStore?.prune && (
				<div
					style={{
						border: "1px solid #ccc",
						marginBottom: "1rem",
						padding: "0.5rem",
					}}
				>
					<div>
						PRUNED SNAPSHOTS:{" "}
						<button onClick={() => managerStore.resetPrune()}>
							RESET
						</button>
					</div>
					<pre>{JSON.stringify(managerStore?.prune || {}, null, 2)}</pre>
				</div>
			)}

			{managerStore.instances.map((instance) => {
				return (
					<div
						style={{
							border: "1px solid #ccc",
							marginBottom: "1rem",
							padding: "0.5rem",
						}}
					>
						<div>
							INSTANCE NAME: <b>{instance.instanceName}</b>
						</div>
						<div>
							IS_ONLINE: <b>{String(instance.isOnline).toUpperCase()}</b>
						</div>
						<div>latestSuccessPingAt: {instance.latestSuccessPingAt}</div>

						<div
							style={{
								border: "1px solid green",
								marginBottom: "1rem",
								padding: "0.5rem",
							}}
						>
							{Object.entries(instance.config?.services || {}).map(
								([serviceName, service]) => (
									<div
										style={{
											border: "1px solid red",
											marginBottom: "1rem",
											padding: "0.5rem",
										}}
									>
										<div>SERVICE: {serviceName}</div>
										<div
											style={{
												border: "1px solid #ccc",
												marginBottom: "1rem",
												padding: "0.5rem",
											}}
										>
											{Object.entries(service || {}).map(([taskName, task]) => {
												return (
													<div
														style={{
															border: "1px solid #ccc",
															marginBottom: "1rem",
															padding: "0.5rem",
														}}
													>
														<div
															style={{
																display: "flex",
																gap: "0.5rem",
															}}
														>
															<div>TASK: {taskName}</div>
															<button
																onClick={() => {
																	if (instance.instanceName) {
																		managerStore.backup(
																			instance.instanceName,
																			serviceName,
																			taskName
																		);
																	}
																}}
															>
																BACKUP{" "}
																{managerStore.backupStatus?.state ===
																	"pending" && "(loading...)"}
															</button>
														</div>
														<div>SOURCE DRIVER: {task.source?.driver}</div>
														<div>
															{Object.entries(task.destinations || {}).map(
																([destinationName, destination], i) => (
																	<div>
																		<div>
																			DESTINATION ({i + 1}): {destinationName}{" "}
																			<b>(driver: {destination.driver})</b>
																		</div>
																		<button
																			onClick={() =>
																				managerStore.fetchSnapshots(
																					instance.instanceName || "",
																					serviceName,
																					taskName,
																					destinationName
																				)
																			}
																		>
																			GET SNAPSHOTS ({managerStore.snapshots?.data?.length})
																			{managerStore.snapshotsStatus?.state ===
																				"pending" && "(loading...)"}
																		</button>

																		<button
																			onClick={() =>
																				managerStore.fetchSnapshotsForPrune(
																					instance.instanceName || "",
																					serviceName,
																					taskName,
																					destinationName
																				)
																			}
																		>
																			GET SNAPSHOTS FOR PRUNE ({managerStore.snapshotsForPrune?.data?.length})
																			{managerStore.snapshotsForPruneStatus
																				?.state === "pending" && "(loading...)"}
																		</button>

																		<button
																			onClick={() =>
																				managerStore.fetchPrune(
																					instance.instanceName || "",
																					serviceName,
																					taskName,
																					destinationName
																				)
																			}
																		>
																			PRUNE ({managerStore.prune?.data?.length})
																			{managerStore.pruneStatus?.state ===
																				"pending" && "(loading...)"}
																		</button>
																	</div>
																)
															)}
														</div>
													</div>
												);
											})}
										</div>
									</div>
								)
							)}
						</div>

						{/* <pre>{JSON.stringify(instance.config || {}, null, 2)}</pre> */}
					</div>
				);
			})}
		</div>
	);
});

export default App;
