import prettyMilliseconds from "pretty-ms";

export interface PrintUptimeProps {
  uptime: number;
}
export const PrintUptime: React.FC<PrintUptimeProps> = ({ uptime }) => {
  const _uptime = !!uptime ? (uptime > 0 ? uptime : uptime * -1) : 0;
  return (
    <>
      {uptime < 0 && "-"}
      {!!_uptime ? prettyMilliseconds(_uptime, { unitCount: 3 }) : "-"}
    </>
  );
};
