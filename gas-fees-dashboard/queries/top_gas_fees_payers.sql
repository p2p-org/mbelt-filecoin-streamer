-- Top 100 Gas Fees Payers
-- Top 100 Gas Fees Payers
-- {"type":"TABLE","name":"Top Gas Fees Payers Table","description":"","options":{"itemsPerPage":10,"columns":[{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"fees_paid","type":"float","displayAs":"number","visible":true,"order":100000,"title":"Fees Paid","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"from","type":"string","displayAs":"string","visible":true,"order":100001,"title":"Address","allowSearch":true,"alignContent":"left","allowHTML":true,"highlightLinks":false}]},"query_id":29}
--
SELECT sum(gas_fees) AS fees_paid,
       "from"
FROM
  (SELECT least((gas_premium * gas_limit) + (blk.parent_base_fee * gas_used), gas_limit * gas_fee_cap) AS gas_fees,
          gas_limit,
          gas_used,
          gas_fee_cap,
          gas_premium,
          parent_base_fee,
          block_cid,
          msg.cid AS msg_cid,
          "from"
   FROM filecoin.messages AS msg
   LEFT JOIN filecoin.blocks blk ON blk.cid = block_cid
   WHERE gas_used > 0 ) AS g
GROUP BY "from"
ORDER BY fees_paid DESC
LIMIT 100;