Import-Module -Name Pester -PassThru

$tests = Get-ChildItem '0*' | ForEach-Object {
  $name=$_.Name;
  $inputFile="$name/input";
  $outputFile="$name/output";
  $arguments=(Get-Content "$name/commandline");
  
  @{ 
    Name = $name;
    InputFile = $inputFile;
    OutputFile = $outputFile;
    Arguments = $arguments;
  };
}

Describe 'sel on PowerShell' {
  It "cat <inputFile> | sel <arguments> Returns <outputFile> (<name>)" -ForEach $tests {
    $expected = Get-Content $outputFile;
    Invoke-Expression "Get-Content $inputFile | ../../../dist/sel $arguments" | Should -Be $expected;
  }
}
