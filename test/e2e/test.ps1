$result = Get-ChildItem 0* | ForEach-Object {
  $name=$_.Name;
  $arguments=(Get-Content "$name/commandline");
  $command="Get-Content $name/input | ../../../dist/sel $arguments";

  Write-Host "${name}: ${command}";
  Compare-Object (Invoke-Expression "$command") (Get-Content "$name/output");
  if ( $? -eq $True ) {
    Write-Host "OK"
    return 0
  } else {
    Write-Host "NG"
    return 1
  }
} | Measure-Object -Sum

if ($result.Sum -eq 0) {
  exit 0
} else {
  exit 1
}
