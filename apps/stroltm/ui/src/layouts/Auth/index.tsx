import { Suspense } from "react";

import { Outlet } from "react-router";

export const Auth = () => {
  return (
    <div>
      <Suspense fallback="loading...">
        <Outlet />
      </Suspense>
    </div>
  );
};
