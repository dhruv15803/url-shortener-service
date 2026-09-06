import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts"
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart"
import { formatBucketFull, type ChartPoint, type Granularity } from "@/lib/analytics"

const chartConfig = {
  clicks: { label: "Clicks", color: "var(--chart-1)" },
} satisfies ChartConfig

type ClicksTrendChartProps = {
  points: ChartPoint[]
  granularity: Granularity
}

export function ClicksTrendChart({ points, granularity }: ClicksTrendChartProps) {
  // Keep the axis readable regardless of how many buckets the range produced.
  const tickInterval = Math.max(0, Math.floor(points.length / 8) - 1)

  return (
    <ChartContainer config={chartConfig} className="h-[260px] w-full">
      <AreaChart data={points} margin={{ left: 4, right: 8, top: 8, bottom: 0 }}>
        <defs>
          <linearGradient id="clicksFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="var(--color-clicks)" stopOpacity={0.35} />
            <stop offset="95%" stopColor="var(--color-clicks)" stopOpacity={0.03} />
          </linearGradient>
        </defs>

        <CartesianGrid vertical={false} strokeDasharray="3 3" />
        <XAxis
          dataKey="label"
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          interval={tickInterval}
          minTickGap={16}
        />
        <YAxis tickLine={false} axisLine={false} tickMargin={8} width={40} allowDecimals={false} />

        <ChartTooltip
          content={
            <ChartTooltipContent
              labelFormatter={(_, payload) => {
                const point = payload?.[0]?.payload as ChartPoint | undefined
                return point ? formatBucketFull(point.bucket, granularity) : ""
              }}
            />
          }
        />

        <Area
          dataKey="clicks"
          type="monotone"
          stroke="var(--color-clicks)"
          strokeWidth={2}
          fill="url(#clicksFill)"
        />
      </AreaChart>
    </ChartContainer>
  )
}
