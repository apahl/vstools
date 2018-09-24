package main

// PdbData holds the molecule string
type PdbData struct {
	Idx int
	PDB string
}

// Row is one table row
type Row struct {
	Idx    int
	ID     string
	Value  float64
	NumHA  int
	LE     float64
	PNGb64 string
	Remark string
}

// Page is the complete HTML page
type Page struct {
	PageNo    string
	Title     string
	Intro     string
	Mols      string
	Rows      string
	LEformula string
	DateStamp string
}

const (
	pdbDataHTML = `<textarea style="display: none;" id="pdbdata-{{.Idx}}">{{.PDB}}</textarea>
`

	rowHTML = `<tr>
<td style="text-align: center;">{{.Idx}}</td>
<td>{{.ID}}</td>
<td>{{printf "%.1f" .Value}}</td>
<td>{{.NumHA}}</td>
<td>{{printf "%.2f" .LE}}</td>
<td><img src="data:image/png;base64,{{.PNGb64}}" alt="Mol"/></td>
<td>
	<div style="height: 300px; width: 300px; position: relative;" class='viewer_3Dmoljs' data-type="pdb" data-element="pdbdata-{{.Idx}}"
		data-backgroundcolor='0x000000' data-style='stick:radius=0.15' data-select1='resn:LIG' data-style1='stick:colorscheme=cyanCarbon' data-select2='resn:HOH' data-style2='sphere:radius~0.4'></div>
</td>
<td style="text-align: left;">{{.Remark}}</td>
</tr>
`

	pageHTML = `<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
	<title>VS{{.PageNo}}</title>
	<script src="http://3Dmol.csb.pitt.edu/build/3Dmol-min.js"></script>
    <style>
        body {
            background-color: #FFFFFF;
            font-family: freesans, arial, verdana, sans-serif;
        }

        th {
            border-collapse: collapse;
            border-width: thin;
            border-style: solid;
			border-color: black;
			background-color: #94CAEF;
            text-align: center;
            font-weight: bold;
        }

        td {
            border-collapse: collapse;
            border-width: thin;
            border-style: solid;
			border-color: black;
			text-align: center;
            padding: 5px;
        }

        table {
            border-collapse: collapse;
            border-width: thin;
            border-style: solid;
            border-color: black;
            border: none;
            background-color: #FFFFFF;
            text-align: left;
        }
    </style>
</head>

<body>
    <h2>{{.Title}}{{.PageNo}}</h2>
    <p>{{.Intro}}</p>
    {{.Mols}}
    <table>
        <tr>
            <th></th>
            <th>Id</th>
            <th>Score</th>
            <th>HA</th>
            <th>LE</th>
            <th>2D Struct</th>
            <th>Binding Mode</th>
            <th style="text-align: left; width: 200px;">Remark</th>
        </tr>
        {{.Rows}}
    </table>
<p><b>Legend:</b><br>
HA: number of heavy atoms (non-hydrogen atoms)<br>
LE: ligand efficiency ({{.LEformula}})<br>
</p>
<p>COMAS {{.DateStamp}}</p>
</body>

</html>
`
)
