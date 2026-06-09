import type { Params } from "react-router";
import { Link as ReactRouterLink } from "react-router-dom";
import type { LinkProps as ReactRouterLinkProps } from "react-router-dom";

import type { ConstantsRouteType } from "boot/routes/constants";

import { toNavigate } from "boot/routes/constants";

export interface LinkProps extends Omit<ReactRouterLinkProps, "to"> {
  href?: string;
  params?: Params;
  styled?: boolean;
  to?: ConstantsRouteType;
}
export const Link: React.FC<LinkProps> = ({ href, params, to, ...props }) => {
  if (href) {
    return <a {...props} href={href} />;
  }

  return <ReactRouterLink {...props} to={to ? toNavigate(to, params) : "/"} />;
};
