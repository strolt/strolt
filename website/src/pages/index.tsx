import { Redirect } from "@docusaurus/router";
import type { ReactNode } from "react";

export default function Docs(): ReactNode {
	return <Redirect to="/docs/intro" />;
}
