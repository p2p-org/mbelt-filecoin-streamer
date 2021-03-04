-- Gas Fees For hour
-- Sums of Gas Fees and premiums paid to miners for every message in an hour of last week
-- {"type":"CHART","name":"Chart","description":"","options":{"globalSeriesType":"area","sortX":true,"legend":{"enabled":true},"yAxis":[{"type":"linear","title":{"text":"FIL"}},{"type":"linear","opposite":true}],"xAxis":{"type":"-","labels":{"enabled":true},"title":{"text":"hour"}},"error_y":{"type":"data","visible":true},"series":{"stacking":null,"error_y":{"type":"data","visible":true}},"seriesOptions":{"premium":{"zIndex":1,"index":0,"type":"area","yAxis":0},"gas_fees":{"zIndex":0,"index":0,"type":"area","yAxis":0}},"valuesOptions":{},"columnMapping":{"premium":"y","hour":"x","gas_fees":"y"},"direction":{"type":"counterclockwise"},"numberFormat":"0,0[.]00000","percentFormat":"0[.]00%","textFormat":"","missingValuesAsZero":true,"showDataLabels":false,"dateTimeFormat":"DD/MM/YY HH:mm","customCode":"// Available variables are x, ys, element, and Plotly\n// Type console.log(x, ys); for more info about x and ys\n// To plot your graph call Plotly.plot(element, ...)\n// Plotly examples and docs: https://plot.ly/javascript/"},"query_id":2}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":3,"sizeY":8,"minSizeX":1,"maxSizeX":6,"minSizeY":5,"maxSizeY":1000,"col":3,"row":22}},"text":"","width":1,"dashboard_id":1,"visualization_id":4}
SELECT sum(gas_fees) * 0.000000000000000001 as gas_fees, sum(premium) * 0.000000000000000001 as premium, "hour"
FROM (
SELECT least((gas_premium * gas_limit) + (blk.parent_base_fee * gas_used), gas_limit * gas_fee_cap) AS gas_fees,
           least((gas_premium * gas_limit) + (blk.parent_base_fee * gas_used), gas_limit * gas_fee_cap) - (blk.parent_base_fee * gas_used) as premium,
           gas_limit,
           gas_used,
           gas_fee_cap,
           gas_premium,
           parent_base_fee,
           date_trunc('hour', msg.block_time) "hour",
           block_cid,
           height as epoch,
           msg.cid as msg_cid
    FROM filecoin.messages as msg
    LEFT JOIN filecoin.blocks blk ON blk.cid = block_cid
    WHERE gas_used > 0 AND height > (SELECT max(height) from filecoin.blocks) - 20160
    GROUP BY "hour",
             epoch,
             block_cid,
             gas_limit,
             gas_used,
             gas_fee_cap,
             gas_premium,
             parent_base_fee,
             msg.cid
) AS g
GROUP BY "hour";