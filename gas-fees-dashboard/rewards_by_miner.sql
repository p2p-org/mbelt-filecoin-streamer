-- Rewards By Miner
--
-- {"type":"CHART","name":"Miners rewards","description":"","options":{"globalSeriesType":"column","sortX":true,"legend":{"enabled":false},"yAxis":[{"type":"linear","title":{"text":"FIL"}},{"type":"linear","opposite":true}],"xAxis":{"type":"-","labels":{"enabled":true},"title":{"text":"miner"}},"error_y":{"type":"data","visible":true},"series":{"stacking":null,"error_y":{"type":"data","visible":true}},"seriesOptions":{"miner_rewards":{"type":"column","yAxis":0,"zIndex":0,"index":0}},"valuesOptions":{},"columnMapping":{"miner":"x","miner_rewards":"y"},"direction":{"type":"counterclockwise"},"numberFormat":"0,0[.]00000","percentFormat":"0[.]00%","textFormat":"","missingValuesAsZero":true,"showDataLabels":false,"dateTimeFormat":"DD/MM/YY HH:mm","customCode":"// Available variables are x, ys, element, and Plotly\n// Type console.log(x, ys); for more info about x and ys\n// To plot your graph call Plotly.plot(element, ...)\n// Plotly examples and docs: https://plot.ly/javascript/"},"query_id":8}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":6,"sizeY":8,"maxSizeY":1000,"maxSizeX":6,"minSizeY":5,"minSizeX":1,"col":0,"row":107}},"text":"","width":1,"dashboard_id":1,"visualization_id":46}
SELECT miner,
       sum(miner_reward) * 0.000000000000000001 AS miner_rewards
FROM (
    SELECT least((gas_premium * gas_limit) + (parent_base_fee * gas_used), gas_limit * gas_fee_cap) - (parent_base_fee * gas_used) AS miner_reward,
           gas_limit,
           gas_used,
           gas_fee_cap,
           gas_premium,
           parent_base_fee,
           miner,
           height AS epoch
    FROM filecoin.messages AS msg
    LEFT JOIN filecoin.blocks blk ON blk.cid = block_cid
    WHERE gas_used > 0
) AS g
GROUP BY miner