import prettyMilliseconds from "pretty-ms";

export interface PrintUptimeProps {
  uptime: number;
}
export const PrintUptime: React.FC<PrintUptimeProps> = ({ uptime }) => {
  return <>{!!uptime ? prettyMilliseconds(uptime, { unitCount: 3 }) : "-"}</>;
};
