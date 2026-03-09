import type { ReactNode } from "react";

import { Redirect } from "@docusaurus/router";

export default function Docs(): ReactNode {
  return <Redirect to="/docs/intro" />;
}
