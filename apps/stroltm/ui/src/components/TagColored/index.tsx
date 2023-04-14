import { Tag } from "antd";
import { getSeededHEXColor, getTagKey } from "utils";


const getColor = (tag: string) => {
  const seed = getTagKey(tag);
	return getSeededHEXColor(seed)
};

export interface TagColoredProps {
  value: string;
}
export const TagColored: React.FC<TagColoredProps> = ({value}) => {
  return <Tag color={getColor(value)}>{value}</Tag>;
};
