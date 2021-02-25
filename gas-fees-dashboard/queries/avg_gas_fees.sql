-- Average Gas Fees For Epoch
-- Average gas fees for each epoch
-- {"type":"CHART","name":"Average Gas Fees For Epoch","description":"","options":{"globalSeriesType":"area","sortX":true,"legend":{"enabled":false},"yAxis":[{"type":"linear","title":{"text":"attoFIL"},"rangeMax":5000000000000},{"type":"linear","opposite":true}],"xAxis":{"type":"-","labels":{"enabled":true},"title":{"text":"epoch"}},"error_y":{"type":"data","visible":true},"series":{"stacking":null,"error_y":{"type":"data","visible":true}},"seriesOptions":{"avg_gas_fees":{"zIndex":0,"index":0,"type":"area","yAxis":0}},"valuesOptions":{},"columnMapping":{"epoch":"x","avg_gas_fees":"y"},"direction":{"type":"counterclockwise"},"numberFormat":"0,0[.]00000","percentFormat":"0[.]00%","textFormat":"","missingValuesAsZero":true,"showDataLabels":false,"dateTimeFormat":"DD/MM/YY HH:mm","customCode":"// Available variables are x, ys, element, and Plotly\n// Type console.log(x, ys); for more info about x and ys\n// To plot your graph call Plotly.plot(element, ...)\n// Plotly examples and docs: https://plot.ly/javascript/"},"query_id":7}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":3,"sizeY":8,"maxSizeY":1000,"maxSizeX":6,"minSizeY":5,"minSizeX":1,"col":3,"row":65}},"text":"","width":1,"dashboard_id":1,"visualization_id":14}
SELECT epoch,
       avg(gas_fees) AS avg_gas_fees
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
    WHERE gas_used > 0 AND method = 0 AND value > 0 AND height > (SELECT max(height) from filecoin.blocks) - 20160
) AS g
GROUP BY epoch