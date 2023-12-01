Get-ChildItem 0* | ForEach-Object {
  $name=$_.Name;
  $arguments=(Get-Content "$name/commandline");
  $command="Get-Content $name/input | ../../../dist/sel $arguments";

  Write-Host "${name}: ${command}";
  Compare-Object (Invoke-Expression "$command") (Get-Content "$name/output");
  if ( $? -eq $True ) {
    Write-Host "OK"
  } else {
    Write-Host "NG"
  }
}
