import { useEffect, useState } from "react";
import * as api from "./api";
import { InstancesGetListResult } from "./api/generated";

function App() {
	const [state, setState] = useState<undefined | InstancesGetListResult>(
		undefined
	);
	const [count, setCount] = useState(0);

	useEffect(() => {
		api.instances.apiInstancesGet().then((data) => {
			setState(data.data);
		});
	}, []);

	return (
		<div>
			{state?.data?.map((el) => {
				return (
					<>
						<div>URL: {el.url}</div>

						{Object.entries(el.operations?.services || {}).map(
							([key, value]) => {
								return (
									<>
										<div>Service: {key}</div>
										<div>
											{Object.entries(value).map(([key1, value1]) => (
												<>
													<div>Task: {key1}</div>
												</>
											))}
										</div>
									</>
								);
							}
						)}

						<pre>{JSON.stringify(el, null, 2)}</pre>
					</>
				);
			})}
		</div>
	);
}

export default App;
