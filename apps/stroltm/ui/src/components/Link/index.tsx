import { toNavigate, ConstantsRouteType } from "boot/routes/constants";

import { Params } from "react-router";
import { Link as ReactRouterLink } from "react-router-dom";
import { LinkProps as ReactRouterLinkProps } from "react-router-dom";

export interface LinkProps extends Omit<ReactRouterLinkProps, "to"> {
  to?: ConstantsRouteType;
  params?: Params;
  href?: string;
  styled?: boolean;
}
export const Link: React.FC<LinkProps> = ({ to, params, href, ...props }) => {
  if (href) {
    return <a {...props} href={href} />;
  }

  return <ReactRouterLink {...props} to={to ? toNavigate(to, params) : "/"} />;
};
