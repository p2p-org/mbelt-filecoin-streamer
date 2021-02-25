-- 100 Most Expensive Messages
-- Table with top 100 messages with biggest amounts of gas paid for them
-- {"type":"TABLE","name":"100 Most Expensive Messages Table","description":"","options":{"itemsPerPage":25,"columns":[{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"gas_fees","type":"float","displayAs":"number","visible":true,"order":100000,"title":"Fee Paid","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"from","type":"string","displayAs":"string","visible":true,"order":100001,"title":"From","allowSearch":true,"alignContent":"left","allowHTML":false,"highlightLinks":false},{"booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"to","type":"string","displayAs":"string","visible":true,"order":100002,"title":"To","allowSearch":false,"alignContent":"left","allowHTML":false,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"method","type":"integer","displayAs":"number","visible":true,"order":100003,"title":"Method","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false}]},"query_id":30}
--
SELECT gas_fees,
       "from",
       "to",
       method
FROM
  (SELECT least((gas_premium * gas_limit) + (blk.parent_base_fee * gas_used), gas_limit * gas_fee_cap) AS gas_fees,
          gas_limit,
          gas_used,
          gas_fee_cap,
          gas_premium,
          parent_base_fee,
          block_cid,
          "from",
          "to",
          METHOD,
          msg.cid AS msg_cid
   FROM filecoin.messages AS msg
   LEFT JOIN filecoin.blocks blk ON blk.cid = block_cid
   WHERE gas_used > 0) AS g
ORDER BY gas_fees DESC
LIMIT 100;