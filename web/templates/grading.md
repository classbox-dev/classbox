{{define "title"}}Grading Policy @ hsecode{{end -}}

# Grading Policy

The 10-point grade for *Theory of Algorithms*
is determined from the total score for passed stdlib tests
by the end of the academic year:

${
  \displaystyle
  \text{Algorithms} = \min \left(
    \frac{\text{Score}}{10}, 10
  \right)
}$


The course grade is a rounded weighted sum of *unrounded* 10-point grades for *Programming* and *Theory of Algorithms*:

${
  \displaystyle
  \text{Course} = \left\lfloor
    0.4 \cdot \text{Algorithms} + 0.6 \cdot \text{Programming} + 0.5
  \right\rfloor
}$

* [Back to main page](..)
