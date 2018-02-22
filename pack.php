<?php
function bin2bstr($input)
// Convert a binary expression (e.g., "100111") into a binary-string
{
  if (!is_string($input)) return null; // Sanity check

  // Pack into a string
  return pack('H*', base_convert($input, 2, 16));
}

function bstr2bin($input)
// Binary representation of a binary-string
{
  if (!is_string($input)) return null; // Sanity check

  // Unpack as a hexadecimal string
  $value = unpack('H*', $input);
  
  // Output binary representation
  return base_convert($value[1], 16, 2);
}

// Returns string(3) "ABC"
var_dump(bin2bstr('01000001 01000010 01000011'));

// Returns string(24) "010000010100001001000011"
var_dump(bstr2bin('ABC'));
?>
