-- Average Gas Fees Components For Epoch
-- Averages for variables involved in gas fees calculation for each epoch
-- {"type":"CHART","name":"Average gas fees components for each epoch","description":"","options":{"globalSeriesType":"area","sortX":true,"legend":{"enabled":true},"yAxis":[{"type":"linear","title":{"text":"attoFIL"}},{"type":"linear","opposite":true}],"xAxis":{"type":"-","labels":{"enabled":true},"title":{"text":"epoch"}},"error_y":{"type":"data","visible":true},"series":{"stacking":null,"error_y":{"type":"data","visible":true}},"seriesOptions":{"avg_gas_premium":{"zIndex":1,"index":0,"type":"area","yAxis":0},"avg_gas_limit":{"zIndex":0,"index":0,"type":"area","yAxis":0},"avg_gas_fee_cap":{"zIndex":3,"index":0,"type":"area","yAxis":0},"avg_gas_used":{"zIndex":2,"index":0,"type":"area","yAxis":0},"avg_parent_base_fee":{"zIndex":4,"index":0,"type":"area","yAxis":0}},"valuesOptions":{},"columnMapping":{"avg_value":"unused","avg_gas_fee_cap":"y","epoch":"x","avg_gas_used":"y","avg_parent_base_fee":"y","avg_gas_fees":"unused","avg_gas_limit":"y","avg_gas_premium":"y"},"direction":{"type":"counterclockwise"},"numberFormat":"0,0[.]00000","percentFormat":"0[.]00%","textFormat":"","missingValuesAsZero":true,"showDataLabels":false,"dateTimeFormat":"DD/MM/YY HH:mm","customCode":"// Available variables are x, ys, element, and Plotly\n// Type console.log(x, ys); for more info about x and ys\n// To plot your graph call Plotly.plot(element, ...)\n// Plotly examples and docs: https://plot.ly/javascript/"},"query_id":6}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":3,"sizeY":8,"minSizeX":1,"maxSizeX":6,"minSizeY":5,"maxSizeY":1000,"col":0,"row":79}},"text":"","width":1,"dashboard_id":1,"visualization_id":50}
SELECT epoch,
       avg(value) AS avg_value,
       avg(gas_fees) AS avg_gas_fees,
       avg(gas_limit) AS avg_gas_limit,
       avg(gas_used) AS avg_gas_used,
       avg(gas_fee_cap) AS avg_gas_fee_cap,
       avg(gas_premium) AS avg_gas_premium,
       avg(parent_base_fee) AS avg_parent_base_fee
FROM (
    SELECT least((gas_premium * gas_limit) + (parent_base_fee * gas_used), gas_limit * gas_fee_cap) AS gas_fees,
           value,
           gas_limit,
           gas_used,
           gas_fee_cap,
           gas_premium,
           parent_base_fee,
           block_cid,
           height AS epoch
    FROM filecoin.messages AS msg
    LEFT JOIN filecoin.blocks blk ON blk.cid = block_cid
    WHERE gas_used > 0 AND method = 0 AND value > 0
) AS g
GROUP BY epoch