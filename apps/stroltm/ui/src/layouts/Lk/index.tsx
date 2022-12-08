import { Suspense } from "react";

import { Outlet } from "react-router";

import { Link } from "components";

import { useStores } from "stores";

export const Lk = () => {
  const { authStore } = useStores();
  return (
    <div>
      <button onClick={() => authStore.logout()}>Logout</button>

      <nav>
        <Link to="instances.main">Instances</Link>
      </nav>
      <Suspense fallback="loading...">
        <Outlet />
      </Suspense>
    </div>
  );
};
