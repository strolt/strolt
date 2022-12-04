import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import { StoreProvider } from "./contexts/storeProvider";
import "./index.css";
import { RootStore } from "./stores/root.store";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
	<React.StrictMode>
		<StoreProvider value={RootStore}>
			<App />
		</StoreProvider>
	</React.StrictMode>
);
