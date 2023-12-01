$result = Get-ChildItem 0* | ForEach-Object {
  $name=$_.Name;
  $arguments=(Get-Content "$name/commandline");
  $command="Get-Content $name/input | ../../../dist/sel $arguments";

  Compare-Object (Invoke-Expression "$command") (Get-Content "$name/output");
  if ( $? -eq $True ) {
    Write-Host "${name}: sel ${arguments} ... OK";
    0;
  } else {
    Write-Host "${name}: sel ${arguments} ... NG";
    1;
  }
} | Measure-Object -Sum

if ($result.Sum -eq 0) {
  exit 0
} else {
  exit 1
}
