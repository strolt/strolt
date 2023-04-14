import { Tag } from "antd";
import { getSeededHEXColor } from "utils";

const getColorSeed = (tag: string) => {
  let seed = tag;

  const one = tag.split(":");
  if (one.length === 2) {
    seed = one[0];
  }

  const two = tag.split("=");
  if (two.length === 2) {
    seed = two[0];
  }

  return seed;
};

const getColor = (tag: string) => {
  const seed = getColorSeed(tag);
	return getSeededHEXColor(seed)
};

export interface TagColoredProps {
  value: string;
}
export const TagColored: React.FC<TagColoredProps> = ({value}) => {
  return <Tag color={getColor(value)}>{value}</Tag>;
};
