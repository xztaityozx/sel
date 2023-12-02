Import-Module -Name Pester -PassThru

$tests = Get-ChildItem '0*' | ForEach-Object {
  $name=$_.Name;
  $inputFile="$name/input";
  $outputFile="$name/output";
  $arguments=(Get-Content -Encoding utf8 "$name/commandline");
  
  @{ 
    Name = $name;
    InputFile = $inputFile;
    OutputFile = $outputFile;
    Arguments = $arguments;
  };
}

$selPath = if ( [System.IO.Path]::Exists("../../../dist/sel.exe") ) {
  "../../../dist/sel.exe"
} else {
  "../../../dist/sel"
}

Describe 'sel on PowerShell' {
  It "sel -f <inputFile> <arguments> Returns <outputFile> (<name>)" -ForEach $tests {
    $expected = Get-Content -Encoding utf8 $outputFile;
    Invoke-Expression "Get-Content -Encoding utf8 $inputFile | $selPath $arguments" | Should -Be $expected;
  }
}
